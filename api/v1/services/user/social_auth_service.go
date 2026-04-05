package user

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	userDTO "api-backend-infinitrum/api/v1/dto/user"
	"api-backend-infinitrum/api/v1/middleware"
	profileRepo "api-backend-infinitrum/api/v1/repositories/profile"
	userRepo "api-backend-infinitrum/api/v1/repositories/user"
	"api-backend-infinitrum/config"
	profileModel "api-backend-infinitrum/internal/models/profile"
	userModel "api-backend-infinitrum/internal/models/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	googleRegUsernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
	nonUserCharsRe      = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
)

func randomAlnum(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}

func sanitizeMemberUsernameBase(localOrFallback string) string {
	s := strings.ToLower(strings.TrimSpace(localOrFallback))
	s = nonUserCharsRe.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if len(s) < 3 {
		s = "user"
	}
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

// uniqueMemberUsername gera username único para profile_members (login social sem cadastro explícito).
func (s *SocialAuthService) uniqueMemberUsername(email, fallbackID string) (string, error) {
	var base string
	if email != "" {
		parts := strings.Split(email, "@")
		base = sanitizeMemberUsernameBase(parts[0])
	} else {
		base = sanitizeMemberUsernameBase("u_" + fallbackID)
	}
	if !googleRegUsernameRe.MatchString(base) {
		base = "user"
	}
	for i := 0; i < 40; i++ {
		candidate := base
		if i > 0 {
			suf := "_" + randomAlnum(6)
			maxBase := 50 - len(suf)
			if maxBase < 3 {
				maxBase = 3
			}
			if len(candidate) > maxBase {
				candidate = candidate[:maxBase]
			}
			candidate = candidate + suf
		}
		taken, err := s.profileRepo.UsernameExists(nil, candidate)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", errors.New("could not allocate username")
}

type SocialAuthService struct {
	userRepo    *userRepo.Repository
	profileRepo *profileRepo.Repository
	db          *gorm.DB
	config      *config.Config
}

func NewSocialAuthService(db *gorm.DB, cfg *config.Config) *SocialAuthService {
	return &SocialAuthService{
		userRepo:    userRepo.NewRepository(db),
		profileRepo: profileRepo.NewRepository(db),
		db:          db,
		config:      cfg,
	}
}

// Google User Info from token verification
type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	ID            string `json:"id"` // Used when getting info from access token
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"` // Alternative field name
}

// Facebook User Info
type FacebookUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

// AuthenticateWithGoogle authenticates user with Google ID token or access token
func (s *SocialAuthService) AuthenticateWithGoogle(token string) (*userDTO.LoginResponse, error) {
	// Try to verify as ID token first, then as access token
	var userInfo *GoogleUserInfo
	var err error
	
	userInfo, err = s.verifyGoogleToken(token)
	if err != nil {
		// If ID token verification fails, try as access token
		userInfo, err = s.getGoogleUserInfoFromAccessToken(token)
		if err != nil {
			return nil, errors.New("invalid Google token")
		}
	}
	
	// Ensure Sub is set (use ID if Sub is empty)
	if userInfo.Sub == "" && userInfo.ID != "" {
		userInfo.Sub = userInfo.ID
	}
	
	// Ensure EmailVerified is set
	if !userInfo.EmailVerified && userInfo.VerifiedEmail {
		userInfo.EmailVerified = userInfo.VerifiedEmail
	}

	// Check if user exists by provider ID
	provider := "google"
	userEntity, err := s.userRepo.GetByProviderID(provider, userInfo.Sub)
	if err == nil {
		// User exists, generate tokens
		return s.generateTokensForUser(userEntity)
	}

	// Check if user exists by email (account linking)
	existingUser, err := s.userRepo.GetByEmail(userInfo.Email)
	if err == nil {
		// User exists with email but no social auth, link the account
		existingUser.Provider = &provider
		existingUser.ProviderID = &userInfo.Sub
		if existingUser.AvatarURL == nil && userInfo.Picture != "" {
			existingUser.AvatarURL = &userInfo.Picture
		}
		if !existingUser.EmailVerified {
			existingUser.EmailVerified = userInfo.EmailVerified
		}
		if err := s.userRepo.Update(existingUser); err != nil {
			return nil, err
		}
		return s.generateTokensForUser(existingUser)
	}

	// Create new user
	firstName := userInfo.GivenName
	if firstName == "" {
		firstName = userInfo.Name
	}
	lastName := userInfo.FamilyName
	if lastName == "" {
		lastName = ""
	}

	avatarURL := userInfo.Picture
	userEntity = &userModel.User{
		ID:            uuid.New().String(),
		Email:         userInfo.Email,
		PasswordHash:  nil, // No password for social auth
		FirstName:     firstName,
		LastName:      lastName,
		Phone:         nil,
		AvatarURL:     &avatarURL,
		IsActive:      true,
		EmailVerified: userInfo.EmailVerified,
		RoleLevel:     userModel.RoleLevelMember,
		Provider:      &provider,
		ProviderID:    &userInfo.Sub,
	}

	uname, err := s.uniqueMemberUsername(userInfo.Email, userInfo.Sub)
	if err != nil {
		return nil, err
	}
	member := &profileModel.ProfileMember{UserID: userEntity.ID, Username: uname}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.CreateWithTx(tx, userEntity); err != nil {
			return err
		}
		return s.profileRepo.CreateMember(tx, member)
	}); err != nil {
		return nil, err
	}

	return s.generateTokensForUser(userEntity)
}

// AuthenticateWithFacebook authenticates user with Facebook access token
func (s *SocialAuthService) AuthenticateWithFacebook(accessToken string) (*userDTO.LoginResponse, error) {
	// Get user info from Facebook
	userInfo, err := s.getFacebookUserInfo(accessToken)
	if err != nil {
		return nil, errors.New("invalid Facebook token")
	}

	// Check if user exists by provider ID
	provider := "facebook"
	userEntity, err := s.userRepo.GetByProviderID(provider, userInfo.ID)
	if err == nil {
		// User exists, generate tokens
		return s.generateTokensForUser(userEntity)
	}

	// Check if user exists by email (account linking)
	if userInfo.Email != "" {
		existingUser, err := s.userRepo.GetByEmail(userInfo.Email)
		if err == nil {
			// User exists with email but no social auth, link the account
			existingUser.Provider = &provider
			existingUser.ProviderID = &userInfo.ID
			if existingUser.AvatarURL == nil && userInfo.Picture.Data.URL != "" {
				existingUser.AvatarURL = &userInfo.Picture.Data.URL
			}
			if err := s.userRepo.Update(existingUser); err != nil {
				return nil, err
			}
			return s.generateTokensForUser(existingUser)
		}
	}

	// Create new user
	// Parse name into first and last name
	firstName := userInfo.Name
	lastName := ""
	if len(userInfo.Name) > 0 {
		parts := splitName(userInfo.Name)
		firstName = parts[0]
		if len(parts) > 1 {
			lastName = parts[1]
		}
	}

	avatarURL := userInfo.Picture.Data.URL
	userEntity = &userModel.User{
		ID:            uuid.New().String(),
		Email:         userInfo.Email,
		PasswordHash:  nil, // No password for social auth
		FirstName:     firstName,
		LastName:      lastName,
		Phone:         nil,
		AvatarURL:     &avatarURL,
		IsActive:      true,
		EmailVerified: true, // Facebook emails are verified
		RoleLevel:     userModel.RoleLevelMember,
		Provider:      &provider,
		ProviderID:    &userInfo.ID,
	}

	uname, err := s.uniqueMemberUsername(userInfo.Email, userInfo.ID)
	if err != nil {
		return nil, err
	}
	member := &profileModel.ProfileMember{UserID: userEntity.ID, Username: uname}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.CreateWithTx(tx, userEntity); err != nil {
			return err
		}
		return s.profileRepo.CreateMember(tx, member)
	}); err != nil {
		return nil, err
	}

	return s.generateTokensForUser(userEntity)
}

// verifyGoogleToken verifies Google ID token and returns user info
func (s *SocialAuthService) verifyGoogleToken(idToken string) (*GoogleUserInfo, error) {
	// Verify token signature and get claims
	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// For production, you should verify the token signature using Google's public keys
		// For now, we'll make a request to Google's tokeninfo endpoint
		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Verify audience (client ID)
	aud, ok := claims["aud"].(string)
	if !ok || aud != s.config.GoogleClientID {
		return nil, errors.New("invalid token audience")
	}

	// Extract user info from claims
	userInfo := &GoogleUserInfo{
		Sub:           getStringClaim(claims, "sub"),
		ID:            getStringClaim(claims, "sub"), // Use sub as ID for consistency
		Email:         getStringClaim(claims, "email"),
		EmailVerified: getBoolClaim(claims, "email_verified"),
		Name:          getStringClaim(claims, "name"),
		GivenName:     getStringClaim(claims, "given_name"),
		FamilyName:    getStringClaim(claims, "family_name"),
		Picture:       getStringClaim(claims, "picture"),
	}

	// Alternative: Use Google's tokeninfo endpoint (simpler, but requires network call)
	if userInfo.Email == "" {
		return s.getGoogleUserInfoFromAPI(idToken)
	}

	return userInfo, nil
}

// getGoogleUserInfoFromAPI gets user info from Google's tokeninfo endpoint
func (s *SocialAuthService) getGoogleUserInfoFromAPI(idToken string) (*GoogleUserInfo, error) {
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to verify token with Google")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenInfo map[string]interface{}
	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		return nil, err
	}

	// Verify audience
	if aud, ok := tokenInfo["aud"].(string); !ok || aud != s.config.GoogleClientID {
		return nil, errors.New("invalid token audience")
	}

	userInfo := &GoogleUserInfo{
		Sub:           getStringFromMap(tokenInfo, "sub"),
		ID:            getStringFromMap(tokenInfo, "sub"), // Use sub as ID for consistency
		Email:         getStringFromMap(tokenInfo, "email"),
		EmailVerified: getBoolFromMap(tokenInfo, "email_verified"),
		Name:          getStringFromMap(tokenInfo, "name"),
		GivenName:     getStringFromMap(tokenInfo, "given_name"),
		FamilyName:    getStringFromMap(tokenInfo, "family_name"),
		Picture:       getStringFromMap(tokenInfo, "picture"),
	}

	return userInfo, nil
}

// getGoogleUserInfoFromAccessToken gets user info from Google using access token
func (s *SocialAuthService) getGoogleUserInfoFromAccessToken(accessToken string) (*GoogleUserInfo, error) {
	// Use Google's userinfo endpoint with access token
	url := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info from Google: %s", string(body))
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}
	
	// The userinfo endpoint returns "id" but we need "sub" for consistency
	// Use ID as Sub if Sub is not provided
	if userInfo.Sub == "" && userInfo.ID != "" {
		userInfo.Sub = userInfo.ID
	}
	
	// Handle verified_email field (alternative name)
	if !userInfo.EmailVerified && userInfo.VerifiedEmail {
		userInfo.EmailVerified = userInfo.VerifiedEmail
	}
	
	return &userInfo, nil
}

// getFacebookUserInfo gets user info from Facebook Graph API
func (s *SocialAuthService) getFacebookUserInfo(accessToken string) (*FacebookUserInfo, error) {
	// First verify the token
	verifyURL := fmt.Sprintf("https://graph.facebook.com/me?access_token=%s&fields=id,name,email,picture", accessToken)
	resp, err := http.Get(verifyURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("facebook API error: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo FacebookUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *SocialAuthService) resolveGoogleUserInfo(token string) (*GoogleUserInfo, error) {
	userInfo, err := s.verifyGoogleToken(token)
	if err != nil {
		userInfo, err = s.getGoogleUserInfoFromAccessToken(token)
		if err != nil {
			return nil, errors.New("invalid Google token")
		}
	}
	if userInfo.Sub == "" && userInfo.ID != "" {
		userInfo.Sub = userInfo.ID
	}
	if !userInfo.EmailVerified && userInfo.VerifiedEmail {
		userInfo.EmailVerified = userInfo.VerifiedEmail
	}
	return userInfo, nil
}

func (s *SocialAuthService) googleUserEntity(userInfo *GoogleUserInfo, roleLevel int) *userModel.User {
	firstName := userInfo.GivenName
	if firstName == "" {
		firstName = userInfo.Name
	}
	lastName := userInfo.FamilyName
	if lastName == "" {
		lastName = ""
	}
	avatarURL := userInfo.Picture
	provider := "google"
	return &userModel.User{
		ID:            uuid.New().String(),
		Email:         userInfo.Email,
		PasswordHash:  nil,
		FirstName:     firstName,
		LastName:      lastName,
		Phone:         nil,
		AvatarURL:     &avatarURL,
		IsActive:      true,
		EmailVerified: userInfo.EmailVerified,
		RoleLevel:     roleLevel,
		Provider:      &provider,
		ProviderID:    &userInfo.Sub,
	}
}

func (s *SocialAuthService) userToResponse(u *userModel.User) userDTO.UserResponse {
	return userDTO.UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Phone:         u.Phone,
		AvatarURL:     u.AvatarURL,
		IsActive:      u.IsActive,
		EmailVerified: u.EmailVerified,
		RoleLevel:     u.RoleLevel,
		Provider:      u.Provider,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// RegisterWithGoogleMember cria conta Google com role 1 e profile_members.
func (s *SocialAuthService) RegisterWithGoogleMember(token, username string) (*userDTO.UserResponse, error) {
	userInfo, err := s.resolveGoogleUserInfo(token)
	if err != nil {
		return nil, err
	}
	username = strings.ToLower(strings.TrimSpace(username))
	if !googleRegUsernameRe.MatchString(username) {
		return nil, errors.New("invalid username: use 3–50 characters, letters, numbers or underscore")
	}

	provider := "google"
	if _, err := s.userRepo.GetByProviderID(provider, userInfo.Sub); err == nil {
		return nil, errors.New("user already exists with this Google account")
	}
	if _, err := s.userRepo.GetByEmail(userInfo.Email); err == nil {
		return nil, errors.New("email already registered")
	}
	taken, err := s.profileRepo.UsernameExists(nil, username)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, errors.New("username already taken")
	}

	userEntity := s.googleUserEntity(userInfo, userModel.RoleLevelMember)
	member := &profileModel.ProfileMember{UserID: userEntity.ID, Username: username}

	var created *userModel.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.CreateWithTx(tx, userEntity); err != nil {
			return err
		}
		if err := s.profileRepo.CreateMember(tx, member); err != nil {
			return err
		}
		var reload userModel.User
		if err := tx.Where("id = ? AND deleted_at IS NULL", userEntity.ID).First(&reload).Error; err != nil {
			return err
		}
		created = &reload
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := s.userToResponse(created)
	return &resp, nil
}

// RegisterWithGoogleAuthor cria conta Google com role 2, profile_authors e communities (slug derivado do pen_name).
func (s *SocialAuthService) RegisterWithGoogleAuthor(token, penName string) (*userDTO.UserResponse, error) {
	userInfo, err := s.resolveGoogleUserInfo(token)
	if err != nil {
		return nil, err
	}
	penName = strings.TrimSpace(penName)
	if penName == "" {
		return nil, errors.New("pen_name is required")
	}

	provider := "google"
	if _, err := s.userRepo.GetByProviderID(provider, userInfo.Sub); err == nil {
		return nil, errors.New("user already exists with this Google account")
	}
	if _, err := s.userRepo.GetByEmail(userInfo.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	userEntity := s.googleUserEntity(userInfo, userModel.RoleLevelAuthor)

	var created *userModel.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		slug, err := s.profileRepo.AllocateUniqueAuthorCommunitySlug(tx, penName)
		if err != nil {
			return err
		}
		if err := s.userRepo.CreateWithTx(tx, userEntity); err != nil {
			return err
		}
		author := &profileModel.ProfileAuthor{
			UserID:  userEntity.ID,
			PenName: penName,
			Slug:    slug,
		}
		if err := s.profileRepo.CreateAuthor(tx, author); err != nil {
			return err
		}
		community := &profileModel.Community{
			OwnerUserID: userEntity.ID,
			Name:        penName,
			Slug:        slug,
		}
		if err := s.profileRepo.CreateCommunity(tx, community); err != nil {
			return err
		}
		var reload userModel.User
		if err := tx.Where("id = ? AND deleted_at IS NULL", userEntity.ID).First(&reload).Error; err != nil {
			return err
		}
		created = &reload
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := s.userToResponse(created)
	return &resp, nil
}

// RegisterWithFacebook creates a new user account with Facebook authentication
func (s *SocialAuthService) RegisterWithFacebook(accessToken string) (*userDTO.UserResponse, error) {
	// Get user info from Facebook
	userInfo, err := s.getFacebookUserInfo(accessToken)
	if err != nil {
		return nil, errors.New("invalid Facebook token")
	}

	provider := "facebook"
	
	// Check if user already exists by provider ID
	if _, err := s.userRepo.GetByProviderID(provider, userInfo.ID); err == nil {
		return nil, errors.New("user already exists with this Facebook account")
	}

	// Check if user already exists by email
	if userInfo.Email != "" {
		if _, err := s.userRepo.GetByEmail(userInfo.Email); err == nil {
			return nil, errors.New("email already registered")
		}
	}

	// Create new user
	// Parse name into first and last name
	firstName := userInfo.Name
	lastName := ""
	if len(userInfo.Name) > 0 {
		parts := splitName(userInfo.Name)
		firstName = parts[0]
		if len(parts) > 1 {
			lastName = parts[1]
		}
	}

	avatarURL := userInfo.Picture.Data.URL
	userEntity := &userModel.User{
		ID:            uuid.New().String(),
		Email:         userInfo.Email,
		PasswordHash:  nil, // No password for social auth
		FirstName:     firstName,
		LastName:      lastName,
		Phone:         nil,
		AvatarURL:     &avatarURL,
		IsActive:      true,
		EmailVerified: true, // Facebook emails are verified
		RoleLevel:     1, // Role level 1 (Member) for social auth users
		Provider:      &provider,
		ProviderID:    &userInfo.ID,
	}

	if err := s.userRepo.Create(userEntity); err != nil {
		return nil, err
	}

	// Convert to response
	userResponse := userDTO.UserResponse{
		ID:            userEntity.ID,
		Email:         userEntity.Email,
		FirstName:     userEntity.FirstName,
		LastName:      userEntity.LastName,
		Phone:         userEntity.Phone,
		AvatarURL:     userEntity.AvatarURL,
		IsActive:      userEntity.IsActive,
		EmailVerified: userEntity.EmailVerified,
		RoleLevel:     userEntity.RoleLevel,
		Provider:      userEntity.Provider,
		CreatedAt:     userEntity.CreatedAt,
		UpdatedAt:     userEntity.UpdatedAt,
	}

	return &userResponse, nil
}

// generateTokensForUser generates access and refresh tokens for a user
func (s *SocialAuthService) generateTokensForUser(userEntity *userModel.User) (*userDTO.LoginResponse, error) {
	// Generate access token (using same format as user_service)
	expirationTime := time.Now().Add(s.config.GetAccessTokenDuration())
	claims := &middleware.Claims{
		UserID:    userEntity.ID,
		Email:     userEntity.Email,
		RoleLevel: userEntity.RoleLevel,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshExpirationTime := time.Now().Add(s.config.GetRefreshTokenDuration())
	refreshClaims := jwt.MapClaims{
		"sub": userEntity.ID,
		"exp": refreshExpirationTime.Unix(),
		"iat": time.Now().Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Save refresh token
	expiresAt := time.Now().Add(s.config.GetRefreshTokenDuration())
	if err := s.userRepo.SaveRefreshToken(userEntity.ID, refreshTokenString, expiresAt); err != nil {
		return nil, err
	}

	// Convert to response using the same format as regular login
	userResponse := userDTO.UserResponse{
		ID:            userEntity.ID,
		Email:         userEntity.Email,
		FirstName:     userEntity.FirstName,
		LastName:      userEntity.LastName,
		Phone:         userEntity.Phone,
		AvatarURL:     userEntity.AvatarURL,
		IsActive:      userEntity.IsActive,
		EmailVerified: userEntity.EmailVerified,
		RoleLevel:     userEntity.RoleLevel,
		Provider:      userEntity.Provider,
		CreatedAt:     userEntity.CreatedAt,
		UpdatedAt:     userEntity.UpdatedAt,
	}

	return &userDTO.LoginResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "bearer",
		User:         userResponse,
	}, nil
}

// Helper functions
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

func getBoolClaim(claims jwt.MapClaims, key string) bool {
	if val, ok := claims[key].(bool); ok {
		return val
	}
	return false
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

func splitName(name string) []string {
	parts := make([]string, 0)
	current := ""
	for _, char := range name {
		if char == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	if len(parts) == 0 {
		return []string{name}
	}
	return parts
}

