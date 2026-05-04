package main

import (
	"context"
	"stvCms/internal/clients"
	"stvCms/internal/config"
	"stvCms/internal/handlers"
	"stvCms/internal/middleware"
	"stvCms/internal/repository"
	"stvCms/internal/services"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func registerRoutes(e *echo.Echo, cfg *config.Config, db *gorm.DB, ctx context.Context) {
	rawRedis := clients.NewRedisClient(ctx, cfg.RedisURL, cfg.RedisAddr, cfg.RedisPassword)
	redisClient := clients.NewRedisWrapper(rawRedis)
	openRouterClient := clients.NewOpenRouter(cfg.OpenRouterAPIKey)
	cloudflareR2 := clients.NewR2Client(ctx, cfg.AccountID, cfg.AccessKeyID, cfg.SecretAccessKey)
	notifRepo := repository.NewNotificationRepository(rawRedis)
	postHandler := handlers.NewPostHandler(ctx, redisClient, openRouterClient, db, cloudflareR2, notifRepo)

	jwtMiddleware := middleware.AuthMiddleware()

	e.GET("/post/image/:filename", postHandler.GetImage)
	e.POST("/post/autoCompleteAI", postHandler.AutoCompleteAI)
	e.GET("/post/getPublic", postHandler.GetPublicPosts)
	e.GET("/post/getPublic/:id", postHandler.GetPublicPostById)

	postGroup := e.Group("/post")
	postGroup.Use(jwtMiddleware)
	postGroup.POST("/create", postHandler.CreatePost)
	postGroup.GET("/getAll", postHandler.GetPosts)
	postGroup.GET("/getPost/:id", postHandler.GetPostById)
	postGroup.PUT("/update", postHandler.UpdatePost)
	postGroup.POST("/uploadImage", postHandler.UploadPostImage)
	postGroup.DELETE("/delete/:id", postHandler.DeletePostById)
	postGroup.POST("/getPost/:filter", postHandler.GetPostByFilter)

	// Admin-only routes for post approval
	adminGroup := e.Group("/admin/post")
	adminGroup.Use(jwtMiddleware, middleware.AdminMiddleware())
	adminGroup.GET("/pending", postHandler.GetPendingPosts)
	adminGroup.PUT("/approve/:id", postHandler.ApprovePost)
	adminGroup.DELETE("/reject/:id", postHandler.RejectPost)

	// Admin notification routes
	notifHandler := handlers.NewNotificationHandler(notifRepo)
	adminNotifGroup := e.Group("/admin/notifications")
	adminNotifGroup.Use(jwtMiddleware, middleware.AdminMiddleware())
	adminNotifGroup.GET("/", notifHandler.GetAll)
	adminNotifGroup.GET("/unread-count", notifHandler.GetUnreadCount)
	adminNotifGroup.PUT("/mark-read/:id", notifHandler.MarkRead)
	adminNotifGroup.PUT("/mark-all-read", notifHandler.MarkAllRead)
	adminNotifGroup.DELETE("/:id", notifHandler.Delete)

	projectHandler := handlers.NewProjectHandler(db)

	e.GET("/project/getPublic", projectHandler.GetPublicProjects)
	e.GET("/project/getPublic/:id", projectHandler.GetProjectById)

	projectGroup := e.Group("/project")
	projectGroup.Use(jwtMiddleware)
	projectGroup.POST("/create", projectHandler.CreateProject)
	projectGroup.GET("/getAll", projectHandler.GetProjects)
	projectGroup.GET("/getProject/:id", projectHandler.GetProjectById)
	projectGroup.PUT("/update", projectHandler.UpdateProject)
	projectGroup.DELETE("/delete/:id", projectHandler.DeleteProjectById)

	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	authGroup := e.Group("/auth")
	authGroup.POST("/google", authHandler.GoogleLogin)
	authGroup.GET("/me", authHandler.Me, jwtMiddleware)

	loginHandler := handlers.NewLoginAndRegisterHandler(authService)
	e.GET("/", loginHandler.Home)
	authGroup.GET("/:provider", loginHandler.SignInWithProvider)
	authGroup.GET("/:provider/callback", loginHandler.CallbackHandler)
	authGroup.GET("/success", loginHandler.Success)
}
