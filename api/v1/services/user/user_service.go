package user

import (
	"errors"
	"time"

	userDTO "api-backend-infinitrum/api/v1/dto/user"
	"api-backend-infinitrum/api/v1/middleware"
	userRepo "api-backend-infinitrum/api/v1/repositories/user"
	"api-backend-infinitrum/config"
	userModel "api-backend-infinitrum/internal/models/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	userRepo *userRepo.Repository
	config   *config.Config
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{
		userRepo: userRepo.NewRepository(db),
		config:   cfg,
	}
}

func (s *Service) Login(email, password string) (*userDTO.LoginResponse, error) {
	// Get user by email
	userEntity, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password (only for non-social auth users)
	if userEntity.PasswordHash == nil {
		return nil, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*userEntity.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(userEntity)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(userEntity)
	if err != nil {
		return nil, err
	}

	// Save refresh token with expiration
	expiresAt := time.Now().Add(s.config.GetRefreshTokenDuration())
	if err := s.userRepo.SaveRefreshToken(userEntity.ID, refreshToken, expiresAt); err != nil {
		return nil, err
	}

	return &userDTO.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		User:         s.toUserResponse(userEntity),
	}, nil
}

func (s *Service) Register(req *userDTO.RegisterRequest) (*userDTO.UserResponse, error) {
	// Check if user already exists
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	passwordHashStr := string(hashedPassword)
	// Create user entity
	userEntity := &userModel.User{
		ID:            uuid.New().String(),
		Email:         req.Email,
		PasswordHash:  &passwordHashStr,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Phone:         req.Phone,
		IsActive:      true,
		EmailVerified: req.EmailVerified,
		RoleLevel:     1, // Role level 1 (Member) for new registrations
	}

	// Save user
	if err := s.userRepo.Create(userEntity); err != nil {
		return nil, err
	}

	// Reload user from database to ensure all fields are populated (including defaults)
	createdUser, err := s.userRepo.GetByID(userEntity.ID)
	if err != nil {
		return nil, err
	}

	userResponse := s.toUserResponse(createdUser)
	return &userResponse, nil
}

func (s *Service) RefreshToken(refreshToken string) (*userDTO.RefreshTokenResponse, error) {
	// Validate refresh token JWT
	userID, err := s.validateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Verify refresh token exists in database and is valid
	session, err := s.userRepo.GetRefreshTokenByToken(userID, refreshToken)
	if err != nil {
		return nil, errors.New("refresh token not found or expired")
	}

	// Check if token is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired token
		s.userRepo.DeleteRefreshToken(userID)
		return nil, errors.New("refresh token expired")
	}

	// Get user
	userEntity, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(userEntity)
	if err != nil {
		return nil, err
	}

	return &userDTO.RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "bearer",
	}, nil
}

func (s *Service) GetUsers(page, limit int) ([]userDTO.UserResponse, error) {
	offset := (page - 1) * limit
	users, err := s.userRepo.GetAll(limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []userDTO.UserResponse
	for _, u := range users {
		responses = append(responses, s.toUserResponse(&u))
	}

	return responses, nil
}

func (s *Service) GetUserByID(userID string) (*userDTO.UserResponse, error) {
	userEntity, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	userResponse := s.toUserResponse(userEntity)
	return &userResponse, nil
}

func (s *Service) UpdateUser(userID string, req *userDTO.UpdateUserRequest) (*userDTO.UserResponse, error) {
	userEntity, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Update fields if provided
	if req.Email != nil {
		userEntity.Email = *req.Email
	}
	if req.FirstName != nil {
		userEntity.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		userEntity.LastName = *req.LastName
	}
	if req.Phone != nil {
		userEntity.Phone = req.Phone
	}
	if req.AvatarURL != nil {
		userEntity.AvatarURL = req.AvatarURL
	}
	if req.EmailVerified != nil {
		userEntity.EmailVerified = *req.EmailVerified
	}
	if req.RoleLevel != nil {
		userEntity.RoleLevel = *req.RoleLevel
	}

	if err := s.userRepo.Update(userEntity); err != nil {
		return nil, err
	}

	userResponse := s.toUserResponse(userEntity)
	return &userResponse, nil
}

func (s *Service) DeleteUser(userID string) error {
	return s.userRepo.SoftDelete(userID)
}

func (s *Service) ChangePassword(userID, currentPassword, newPassword string) error {
	userEntity, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify current password (only for non-social auth users)
	if userEntity.PasswordHash == nil {
		return errors.New("password change not available for social authentication users")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*userEntity.PasswordHash), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(userID, string(hashedPassword))
}

func (s *Service) Logout(userID string) error {
	// Delete all refresh tokens for the user
	return s.userRepo.DeleteRefreshToken(userID)
}

func (s *Service) ForgotPassword(email string) error {
	// Get user by email
	userEntity, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Don't reveal if user exists or not (security best practice)
		return nil
	}

	// Check if user has password (not social auth only)
	if userEntity.PasswordHash == nil {
		// User is social auth only, can't reset password
		return nil
	}

	// Generate reset token
	resetToken := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour

	passwordReset := &userModel.PasswordReset{
		UserID:    userEntity.ID,
		Token:     resetToken,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	if err := s.userRepo.CreatePasswordResetToken(passwordReset); err != nil {
		return err
	}

	// TODO: Send email with reset token
	// For now, we'll just log it (in production, send email)
	// Example: sendResetPasswordEmail(userEntity.Email, resetToken)

	return nil
}

func (s *Service) ResetPassword(token, newPassword string) error {
	// Get reset token
	resetToken, err := s.userRepo.GetPasswordResetToken(token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return errors.New("reset token has expired")
	}

	// Check if token was already used
	if resetToken.Used {
		return errors.New("reset token has already been used")
	}

	// Get user
	userEntity, err := s.userRepo.GetByID(resetToken.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	if err := s.userRepo.UpdatePassword(userEntity.ID, string(hashedPassword)); err != nil {
		return err
	}

	// Mark token as used
	if err := s.userRepo.MarkPasswordResetTokenAsUsed(token); err != nil {
		return err
	}

	// Invalidate all refresh tokens for security
	s.userRepo.DeleteRefreshToken(userEntity.ID)

	return nil
}

// Helper methods
func (s *Service) generateAccessToken(user *userModel.User) (string, error) {
	expirationTime := time.Now().Add(s.config.GetAccessTokenDuration())
	claims := &middleware.Claims{
		UserID:    user.ID,
		Email:     user.Email,
		RoleLevel: user.RoleLevel,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) generateRefreshToken(user *userModel.User) (string, error) {
	expirationTime := time.Now().Add(s.config.GetRefreshTokenDuration())
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": expirationTime.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) validateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("invalid user ID in token")
	}

	return userID, nil
}

func (s *Service) toUserResponse(user *userModel.User) userDTO.UserResponse {
	return userDTO.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Phone:         user.Phone,
		AvatarURL:     user.AvatarURL,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		RoleLevel:     user.RoleLevel,
		Provider:      user.Provider,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
