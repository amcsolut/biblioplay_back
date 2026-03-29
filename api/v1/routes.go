package v1

import (
	"api-backend-infinitrum/api/v1/handlers/user"
	"api-backend-infinitrum/api/v1/middleware"
	"api-backend-infinitrum/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// API v1 group
	v1 := router.Group("/api/v1")

	// Initialize handlers
	userHandler := user.NewHandler(db, cfg)

	// Auth routes (no middleware required)
	auth := v1.Group("/auth")
	{
		auth.POST("/login", userHandler.Login)
		auth.POST("/refresh", userHandler.RefreshToken)
		auth.POST("/register", userHandler.Register)
		auth.POST("/google", userHandler.GoogleAuth)
		auth.POST("/facebook", userHandler.FacebookAuth)
		auth.POST("/register/google", userHandler.RegisterWithGoogle)
		auth.POST("/register/facebook", userHandler.RegisterWithFacebook)
		auth.POST("/forgot-password", userHandler.ForgotPassword)
		auth.POST("/reset-password", userHandler.ResetPassword)
	}

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// Auth routes (protected)
		auth := protected.Group("/auth")
		{
			auth.GET("/me", userHandler.GetMe)
			auth.POST("/logout", userHandler.Logout)
		}

		// User routes
		users := protected.Group("/users")
		{
			users.GET("", userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireAdmin(), userHandler.DeleteUser)
			users.POST("/change-password", userHandler.ChangePassword)
		}

		// User Invitations routes
		invitations := protected.Group("/invitations")
		{
			invitations.GET("", userHandler.GetInvitations)
			invitations.POST("", userHandler.CreateInvitation)
			invitations.GET("/:id", userHandler.GetInvitation)
			invitations.PUT("/:id/accept", userHandler.AcceptInvitation)
			invitations.DELETE("/:id", userHandler.DeleteInvitation)
		}
	}
}
