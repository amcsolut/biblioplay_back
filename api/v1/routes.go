package v1

import (
	cataloghandler "api-backend-infinitrum/api/v1/handlers/catalog"
	commercehandler "api-backend-infinitrum/api/v1/handlers/commerce"
	communityhandler "api-backend-infinitrum/api/v1/handlers/community"
	feedhandler "api-backend-infinitrum/api/v1/handlers/feed"
	libraryhandler "api-backend-infinitrum/api/v1/handlers/library"
	"api-backend-infinitrum/api/v1/handlers/user"
	"api-backend-infinitrum/api/v1/middleware"
	"api-backend-infinitrum/config"
	usermodel "api-backend-infinitrum/internal/models/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// API v1 group
	v1 := router.Group("/api/v1")

	// Initialize handlers
	userHandler := user.NewHandler(db, cfg)
	communityH := communityhandler.NewHandler(db)
	catalogH := cataloghandler.NewHandler(db)
	libraryH := libraryhandler.NewHandler(db)
	commerceH := commercehandler.NewHandler(db)
	feedH := feedhandler.NewHandler(db)

	// Auth routes (no middleware required)
	auth := v1.Group("/auth")
	{
		auth.POST("/login", userHandler.Login)
		auth.POST("/refresh", userHandler.RefreshToken)
		auth.POST("/register/member", userHandler.RegisterMember)
		auth.POST("/register/author", userHandler.RegisterAuthor)
		auth.POST("/google", userHandler.GoogleAuth)
		auth.POST("/register/google/member", userHandler.RegisterWithGoogleMember)
		auth.POST("/register/google/author", userHandler.RegisterWithGoogleAuthor)
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

		// Biblioteca do leitor (itens polimórficos) + compras avulsas
		me := protected.Group("/me")
		{
			me.GET("/library", libraryH.ListMyLibrary)
			me.POST("/library", libraryH.AddFreeToLibrary)
			me.DELETE("/library/:itemType/:itemId", libraryH.RemoveFromLibrary)
		}

		protected.POST("/purchases", commerceH.CreatePurchase)

		authors := protected.Group("/authors")
		authors.Use(middleware.RequireRoleLevel(usermodel.RoleLevelAuthor))
		{
			authors.POST("/grants", libraryH.AuthorGrant)
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

		// Communities + feed (posts, mídia, comentários, replies)
		communities := protected.Group("/communities")
		{
			// Rotas aninhadas antes de GET /:id (detalhe da comunidade)
			communities.GET("/:id/posts/:postId/media/:mediaId", feedH.GetPostMedia)
			communities.GET("/:id/posts/:postId/media", feedH.ListPostMedia)
			communities.POST("/:id/posts/:postId/media", feedH.CreatePostMedia)
			communities.PUT("/:id/posts/:postId/media/:mediaId", feedH.UpdatePostMedia)
			communities.DELETE("/:id/posts/:postId/media/:mediaId", feedH.DeletePostMedia)

			communities.GET("/:id/posts/:postId/comments/:commentId/replies/:replyId", feedH.GetReply)
			communities.GET("/:id/posts/:postId/comments/:commentId/replies", feedH.ListReplies)
			communities.POST("/:id/posts/:postId/comments/:commentId/replies", feedH.CreateReply)
			communities.PUT("/:id/posts/:postId/comments/:commentId/replies/:replyId", feedH.UpdateReply)
			communities.DELETE("/:id/posts/:postId/comments/:commentId/replies/:replyId", feedH.DeleteReply)

			communities.GET("/:id/posts/:postId/comments/:commentId", feedH.GetComment)
			communities.GET("/:id/posts/:postId/comments", feedH.ListComments)
			communities.POST("/:id/posts/:postId/comments", feedH.CreateComment)
			communities.PUT("/:id/posts/:postId/comments/:commentId", feedH.UpdateComment)
			communities.DELETE("/:id/posts/:postId/comments/:commentId", feedH.DeleteComment)

			communities.GET("/:id/posts/:postId", feedH.GetPost)
			communities.PUT("/:id/posts/:postId", feedH.UpdatePost)
			communities.DELETE("/:id/posts/:postId", feedH.DeletePost)
			communities.GET("/:id/posts", feedH.ListPosts)
			communities.POST("/:id/posts", feedH.CreatePost)

			communities.GET("", communityH.List)
			communities.POST("", communityH.Create)
			communities.GET("/:id", communityH.Get)
			communities.PUT("/:id", communityH.Update)
			communities.DELETE("/:id", communityH.Delete)
		}

		// Reações (alvo polimórfico)
		feedRoutes := protected.Group("/feed")
		{
			feedRoutes.GET("/reactions", feedH.ListReactions)
			feedRoutes.POST("/reactions", feedH.UpsertReaction)
			feedRoutes.DELETE("/reactions/:reactionId", feedH.DeleteReaction)
		}

		// Catálogo: coleções (agrupamento de obras)
		collections := protected.Group("/catalog/collections")
		{
			collections.GET("", catalogH.ListCollections)
			collections.POST("", catalogH.CreateCollection)
			collections.GET("/:collectionId", catalogH.GetCollection)
			collections.PUT("/:collectionId", catalogH.UpdateCollection)
			collections.DELETE("/:collectionId", catalogH.DeleteCollection)
			collections.PUT("/:collectionId/books", catalogH.ReplaceCollectionBooks)
		}

		// Catálogo: obras e capítulos (autor = usuário autenticado)
		books := protected.Group("/catalog/books")
		{
			books.GET("", catalogH.ListBooks)
			books.POST("", catalogH.CreateBook)

			books.GET("/:bookId/ebook-chapters/:chapterId", catalogH.GetEbookChapter)
			books.GET("/:bookId/ebook-chapters", catalogH.ListEbookChapters)
			books.POST("/:bookId/ebook-chapters", catalogH.CreateEbookChapter)
			books.PUT("/:bookId/ebook-chapters/:chapterId", catalogH.UpdateEbookChapter)
			books.DELETE("/:bookId/ebook-chapters/:chapterId", catalogH.DeleteEbookChapter)

			books.GET("/:bookId/audiobook-chapters/:chapterId", catalogH.GetAudiobookChapter)
			books.GET("/:bookId/audiobook-chapters", catalogH.ListAudiobookChapters)
			books.POST("/:bookId/audiobook-chapters", catalogH.CreateAudiobookChapter)
			books.PUT("/:bookId/audiobook-chapters/:chapterId", catalogH.UpdateAudiobookChapter)
			books.DELETE("/:bookId/audiobook-chapters/:chapterId", catalogH.DeleteAudiobookChapter)

			books.GET("/:bookId", catalogH.GetBook)
			books.PUT("/:bookId", catalogH.UpdateBook)
			books.DELETE("/:bookId", catalogH.DeleteBook)
		}
	}
}
