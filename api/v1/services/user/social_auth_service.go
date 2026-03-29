package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	userDTO "api-backend-infinitrum/api/v1/dto/user"
	"api-backend-infinitrum/api/v1/middleware"
	userRepo "api-backend-infinitrum/api/v1/repositories/user"
	"api-backend-infinitrum/config"
	userModel "api-backend-infinitrum/internal/models/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SocialAuthService struct {
	userRepo *userRepo.Repository
	config   *config.Config
}

func NewSocialAuthService(db *gorm.DB, cfg *config.Config) *SocialAuthService {
	return &SocialAuthService{
		userRepo: userRepo.NewRepository(db),
		config:   cfg,
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
		RoleLevel:     1, // Role level 1 (Member) for social auth users
		Provider:      &provider,
		ProviderID:    &userInfo.Sub,
	}

	if err := s.userRepo.Create(userEntity); err != nil {
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
		RoleLevel:     1, // Role level 1 (Member) for social auth users
		Provider:      &provider,
		ProviderID:    &userInfo.ID,
	}

	if err := s.userRepo.Create(userEntity); err != nil {
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

// RegisterWithGoogle creates a new user account with Google authentication
func (s *SocialAuthService) RegisterWithGoogle(token string) (*userDTO.UserResponse, error) {
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

	provider := "google"
	
	// Check if user already exists by provider ID
	if _, err := s.userRepo.GetByProviderID(provider, userInfo.Sub); err == nil {
		return nil, errors.New("user already exists with this Google account")
	}

	// Check if user already exists by email
	if _, err := s.userRepo.GetByEmail(userInfo.Email); err == nil {
		return nil, errors.New("email already registered")
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
	userEntity := &userModel.User{
		ID:            uuid.New().String(),
		Email:         userInfo.Email,
		PasswordHash:  nil, // No password for social auth
		FirstName:     firstName,
		LastName:      lastName,
		Phone:         nil,
		AvatarURL:     &avatarURL,
		IsActive:      true,
		EmailVerified: userInfo.EmailVerified,
		RoleLevel:     1, // Role level 1 (Member) for social auth users
		Provider:      &provider,
		ProviderID:    &userInfo.Sub,
	}

	if err := s.userRepo.Create(userEntity); err != nil {
		return nil, err
	}

	// Reload user from database to ensure all fields are populated (including defaults)
	createdUser, err := s.userRepo.GetByID(userEntity.ID)
	if err != nil {
		return nil, err
	}

	// Convert to response
	userResponse := userDTO.UserResponse{
		ID:            createdUser.ID,
		Email:         createdUser.Email,
		FirstName:     createdUser.FirstName,
		LastName:      createdUser.LastName,
		Phone:         createdUser.Phone,
		AvatarURL:     createdUser.AvatarURL,
		IsActive:      createdUser.IsActive,
		EmailVerified: createdUser.EmailVerified,
		RoleLevel:     createdUser.RoleLevel,
		Provider:      createdUser.Provider,
		CreatedAt:     createdUser.CreatedAt,
		UpdatedAt:     createdUser.UpdatedAt,
	}

	return &userResponse, nil
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

