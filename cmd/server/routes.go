package main

import (
	"context"
	"stvCms/internal/clients"
	"stvCms/internal/config"
	"stvCms/internal/handlers"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func registerRoutes(e *echo.Echo, cfg *config.Config, db *gorm.DB, ctx context.Context) {
	redisClient := clients.NewRedisClient(ctx, cfg.RedisURL, cfg.RedisAddr, cfg.RedisPassword)
	openRouterClient := clients.NewOpenRouter(cfg.OpenRouterAPIKey)
	postHandler := handlers.NewPostHandler(ctx, *redisClient, openRouterClient, db)

	postGroup := e.Group("/post")
	postGroup.POST("/create", postHandler.CreatePost)
	postGroup.GET("/getAll", postHandler.GetPosts)
	postGroup.GET("/getPost/:id", postHandler.GetPostById)
	postGroup.PUT("/update", postHandler.UpdatePost)
	postGroup.POST("/uploadImage", postHandler.UploadPostImage)
	postGroup.GET("/image/:filename", postHandler.GetImage)
	postGroup.DELETE("/delete/:id", postHandler.DeletePostById)
	postGroup.POST("/getPost/:filter", postHandler.GetPostByFilter)
	postGroup.POST("/autoCompleteAI", postHandler.AutoCompleteAI)

	authHandler := handlers.NewLoginAndRegisterHandler()
	authGroup := e.Group("/auth")
	authGroup.GET("/", authHandler.Home)
	authGroup.GET("/:provider", authHandler.SignInWithProvider)
	authGroup.GET("/:provider/callback", authHandler.CallbackHandler)
	authGroup.GET("/success", authHandler.Success)
}
