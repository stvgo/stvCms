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
	redisClient := clients.NewRedisWrapper(clients.NewRedisClient(ctx, cfg.RedisURL, cfg.RedisAddr, cfg.RedisPassword))
	openRouterClient := clients.NewOpenRouter(cfg.OpenRouterAPIKey)
	cloudflareR2 := clients.NewR2Client(ctx, cfg.AccountID, cfg.AccessKeyID, cfg.SecretAccessKey)
	postHandler := handlers.NewPostHandler(ctx, redisClient, openRouterClient, db, cloudflareR2)

	jwtMiddleware := middleware.AuthMiddleware()

	e.GET("/post/image/:filename", postHandler.GetImage)
	e.POST("/post/autoCompleteAI", postHandler.AutoCompleteAI)

	postGroup := e.Group("/post")
	postGroup.Use(jwtMiddleware)
	postGroup.POST("/create", postHandler.CreatePost)
	postGroup.GET("/getAll", postHandler.GetPosts)
	postGroup.GET("/getPost/:id", postHandler.GetPostById)
	postGroup.PUT("/update", postHandler.UpdatePost)
	postGroup.POST("/uploadImage", postHandler.UploadPostImage)
	postGroup.DELETE("/delete/:id", postHandler.DeletePostById)
	postGroup.POST("/getPost/:filter", postHandler.GetPostByFilter)

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
