package user

import (
	"net/http"
	"strconv"

	userDTO "api-backend-infinitrum/api/v1/dto/user"
	"api-backend-infinitrum/api/v1/middleware"
	userService "api-backend-infinitrum/api/v1/services/user"
	"api-backend-infinitrum/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	userService     *userService.Service
	socialAuthService *userService.SocialAuthService
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		userService:      userService.NewService(db, cfg),
		socialAuthService: userService.NewSocialAuthService(db, cfg),
	}
}

// @Summary User login
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body userDTO.LoginRequest true "Login credentials"
// @Success 200 {object} userDTO.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req userDTO.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func bindRegisterJSON(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		var validationErrors []string
		if validationErr, ok := err.(interface{ Unwrap() []error }); ok {
			for _, e := range validationErr.Unwrap() {
				validationErrors = append(validationErrors, e.Error())
			}
		}
		response := gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		}
		if len(validationErrors) > 0 {
			response["validation_errors"] = validationErrors
		}
		c.JSON(http.StatusBadRequest, response)
		return false
	}
	return true
}

// @Summary Register member (reader/listener)
// @Description Cria conta com role_level 1 e registro em profile_members
// @Tags auth
// @Accept json
// @Produce json
// @Param request body userDTO.RegisterMemberRequest true "Member registration"
// @Success 201 {object} userDTO.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/register/member [post]
func (h *Handler) RegisterMember(c *gin.Context) {
	var req userDTO.RegisterMemberRequest
	if !bindRegisterJSON(c, &req) {
		return
	}
	userResponse, err := h.userService.RegisterMember(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, userResponse)
}

// @Summary Register author
// @Description Cria conta com role_level 2, profile_authors e communities (slug único derivado do pen_name)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body userDTO.RegisterAuthorRequest true "Author registration"
// @Success 201 {object} userDTO.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/register/author [post]
func (h *Handler) RegisterAuthor(c *gin.Context) {
	var req userDTO.RegisterAuthorRequest
	if !bindRegisterJSON(c, &req) {
		return
	}
	userResponse, err := h.userService.RegisterAuthor(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, userResponse)
}

// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body userDTO.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} userDTO.RefreshTokenResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req userDTO.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get current user
// @Description Get information about the currently authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} userDTO.UserResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.userService.GetUserByID(currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Logout user
// @Description Logout user and invalidate refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	currentUserID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err := h.userService.Logout(currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// @Summary Get users
// @Description Get list of users with pagination
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} userDTO.UserResponse
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users, err := h.userService.GetUsers(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// @Summary Get user by ID
// @Description Get user information by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} userDTO.UserResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	userResponse, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body userDTO.UpdateUserRequest true "User update data"
// @Success 200 {object} userDTO.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID, _ := middleware.GetCurrentUserID(c)

	// Users can only update their own profile (unless admin)
	if userID != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own profile"})
		return
	}

	var req userDTO.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userResponse, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userResponse)
}

// @Summary Delete user
// @Description Soft delete user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	err := h.userService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Change password
// @Description Change user password
// @Tags users
// @Accept json
// @Produce json
// @Param request body userDTO.ChangePasswordRequest true "Password change data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/users/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	currentUserID, _ := middleware.GetCurrentUserID(c)

	var req userDTO.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userService.ChangePassword(currentUserID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// @Summary Listar convites
// @Description Reservado — ainda não implementado
// @Tags invitations
// @Produce json
// @Success 501 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/invitations [get]
func (h *Handler) GetInvitations(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
}

// @Summary Criar convite
// @Description Reservado — ainda não implementado
// @Tags invitations
// @Produce json
// @Success 501 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/invitations [post]
func (h *Handler) CreateInvitation(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
}

// @Summary Obter convite por ID
// @Description Reservado — ainda não implementado
// @Tags invitations
// @Produce json
// @Param id path string true "ID do convite"
// @Success 501 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/invitations/{id} [get]
func (h *Handler) GetInvitation(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
}

// @Summary Aceitar convite
// @Description Reservado — ainda não implementado
// @Tags invitations
// @Produce json
// @Param id path string true "ID do convite"
// @Success 501 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/invitations/{id}/accept [put]
func (h *Handler) AcceptInvitation(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
}

// @Summary Excluir convite
// @Description Reservado — ainda não implementado
// @Tags invitations
// @Produce json
// @Param id path string true "ID do convite"
// @Success 501 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/v1/invitations/{id} [delete]
func (h *Handler) DeleteInvitation(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
}
// @Summary Request password reset
// @Description Send password reset token to user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body userDTO.ForgotPasswordRequest true "User email"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req userDTO.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userService.ForgotPassword(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password reset request"})
		return
	}

	// Always return success (don't reveal if email exists)
	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// @Summary Reset password
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body userDTO.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req userDTO.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.userService.ResetPassword(req.Token, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}

// @Summary Authenticate with Google
// @Description Authenticate user with Google ID token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body userDTO.GoogleAuthRequest true "Google ID token"
// @Success 200 {object} userDTO.LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/google [post]
func (h *Handler) GoogleAuth(c *gin.Context) {
	var req userDTO.GoogleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.socialAuthService.AuthenticateWithGoogle(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Register with Google (member)
// @Description Cria conta Google com role 1 e profile_members
// @Tags auth
// @Accept json
// @Produce json
// @Param request body userDTO.GoogleRegisterMemberRequest true "Google token + username"
// @Success 201 {object} userDTO.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/v1/auth/register/google/member [post]
func (h *Handler) RegisterWithGoogleMember(c *gin.Context) {
	var req userDTO.GoogleRegisterMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userResponse, err := h.socialAuthService.RegisterWithGoogleMember(req.Token, req.Username)
	if err != nil {
		if err.Error() == "user already exists with this Google account" || err.Error() == "email already registered" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, userResponse)
}

// @Summary Register with Google (author)
// @Description Cria conta Google com role 2, profile_authors e communities (slug a partir do pen_name)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body userDTO.GoogleRegisterAuthorRequest true "Google token + pen_name"
// @Success 201 {object} userDTO.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /api/v1/auth/register/google/author [post]
func (h *Handler) RegisterWithGoogleAuthor(c *gin.Context) {
	var req userDTO.GoogleRegisterAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userResponse, err := h.socialAuthService.RegisterWithGoogleAuthor(req.Token, req.PenName)
	if err != nil {
		if err.Error() == "user already exists with this Google account" || err.Error() == "email already registered" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, userResponse)
}
