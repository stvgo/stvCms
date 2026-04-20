package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"stvCms/internal/clients"
	"stvCms/internal/config"
	"stvCms/internal/handlers"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func main() {
	loadEnv()
	initAuth()
	startDatabase()
	startServer()
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from system environment")
	}
}

func initAuth() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	callbackURL := os.Getenv("CLIENT_CALLBACK_URL")

	if clientID == "" || clientSecret == "" || callbackURL == "" {
		slog.Error("Warning: Google OAuth not configured, skipping auth setup")
		return
	}

	goth.UseProviders(
		google.New(clientID, clientSecret, callbackURL),
	)

	key := os.Getenv("SESSION_SECRET")
	if key == "" {
		key = "default-secret-key-change-this"
		log.Println("Warning: SESSION_SECRET not set, using default key")
	}
	gothic.Store = sessions.NewCookieStore([]byte(key))
}

func startServer() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	ctx := context.Background()
	redisClient := clients.NewRedisClient(ctx, redisAddr, redisPassword)

	openRouterClient := clients.NewOpenRouter()
	postHandler := handlers.NewPostHandler(ctx, *redisClient, openRouterClient)

	postGroup := e.Group("/post")
	postGroup.POST("/create", postHandler.CreatePost)
	postGroup.GET("/getAll", postHandler.GetPosts)
	postGroup.GET("/getPost/:id", postHandler.GetPostById)
	postGroup.PUT("/update", postHandler.UpdatePost)
	postGroup.POST("/uploadImage", postHandler.UploadPostImage)
	postGroup.GET("/image/:filename", postHandler.GetImage)
	postGroup.DELETE("/delete/:id", postHandler.DeletePostById)
	postGroup.POST("/getPost/:filter", postHandler.GetPostByFilter)
	postGroup.POST("/genTextAI", postHandler.GetTextAI)

	authHandler := handlers.NewLoginAndRegisterHandler()
	authGroup := e.Group("/auth")
	authGroup.GET("/", authHandler.Home)
	authGroup.GET("/:provider", authHandler.SignInWithProvider)
	authGroup.GET("/:provider/callback", authHandler.CallbackHandler)
	authGroup.GET("/success", authHandler.Success)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	err := e.Start("0.0.0.0:" + port)
	if err != nil {
		panic(err)
	}
}

func startDatabase() {
	config.Init()
}
