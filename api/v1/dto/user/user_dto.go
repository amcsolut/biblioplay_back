package user

import (
	"time"
)

// Request DTOs
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterMemberRequest cadastro de leitor/ouvinte (role_level 1 + profile_members).
type RegisterMemberRequest struct {
	Email         string  `json:"email" binding:"required,email"`
	Password      string  `json:"password" binding:"required,min=6"`
	FirstName     string  `json:"first_name" binding:"required,min=2"`
	LastName      string  `json:"last_name" binding:"required,min=2"`
	Phone         *string `json:"phone"`
	EmailVerified bool    `json:"email_verified"`
	Username      string  `json:"username" binding:"required,min=3,max=50"`
}

// RegisterAuthorRequest cadastro de autor (role_level 2 + profile_authors + communities).
// Slug do autor e da comunidade é gerado a partir de pen_name (minúsculas, números, hífens).
type RegisterAuthorRequest struct {
	Email         string  `json:"email" binding:"required,email"`
	Password      string  `json:"password" binding:"required,min=6"`
	FirstName     string  `json:"first_name" binding:"required,min=2"`
	LastName      string  `json:"last_name" binding:"required,min=2"`
	Phone         *string `json:"phone"`
	EmailVerified bool    `json:"email_verified"`
	PenName       string  `json:"pen_name" binding:"required,min=1,max=200"`
}

type UpdateUserRequest struct {
	Email         *string `json:"email" binding:"omitempty,email"`
	FirstName     *string `json:"first_name" binding:"omitempty,min=2"`
	LastName      *string `json:"last_name" binding:"omitempty,min=2"`
	Phone         *string `json:"phone"`
	AvatarURL     *string `json:"avatar_url"`
	EmailVerified *bool   `json:"email_verified"`
	RoleLevel     *int    `json:"role_level" binding:"omitempty,min=1,max=9"` // Apenas admins podem alterar
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// Social Authentication DTOs
type SocialAuthRequest struct {
	Token string `json:"token" binding:"required"` // OAuth access token from provider
}

type GoogleAuthRequest struct {
	Token string `json:"token" binding:"required"` // Google ID token ou access token
}

// GoogleRegisterMemberRequest POST /auth/register/google/member
type GoogleRegisterMemberRequest struct {
	Token    string `json:"token" binding:"required"`
	Username string `json:"username" binding:"required,min=3,max=50"`
}

// GoogleRegisterAuthorRequest POST /auth/register/google/author
type GoogleRegisterAuthorRequest struct {
	Token   string `json:"token" binding:"required"`
	PenName string `json:"pen_name" binding:"required,min=1,max=200"`
}

type FacebookAuthRequest struct {
	Token string `json:"token" binding:"required"` // Facebook access token
}

// Response DTOs
type UserResponse struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Phone         *string   `json:"phone"`
	AvatarURL     *string   `json:"avatar_url"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	RoleLevel     int       `json:"role_level"`
	Provider      *string   `json:"provider"` // google, facebook, or null
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// Invitation DTOs
type CreateInvitationRequest struct {
	Email          string `json:"email" binding:"required,email"`
	OrganizationID string `json:"organization_id" binding:"required"`
}

type InvitationResponse struct {
	ID             string     `json:"id"`
	Email          string     `json:"email"`
	OrganizationID string     `json:"organization_id"`
	InvitedBy      string     `json:"invited_by"`
	Status         string     `json:"status"`
	ExpiresAt      time.Time  `json:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

