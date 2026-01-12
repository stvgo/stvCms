package main

import (
	"log"
	"os"
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
		panic("Error loading .env file")
	}
}

func initAuth() {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	callbackURL := os.Getenv("CLIENT_CALLBACK_URL")

	if clientID == "" || clientSecret == "" || callbackURL == "" {
		log.Fatal("Environment variables (CLIENT_ID, CLIENT_SECRET, CLIENT_CALLBACK_URL) are required")
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

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Handlers
	postHandler := handlers.NewPostHandler()
	//login := handlers.NewLoginAndRegisterHandler()

	// Post routes
	postGroup := e.Group("/post")
	postGroup.POST("/create", postHandler.CreatePost)
	postGroup.GET("/getAll", postHandler.GetPosts)
	postGroup.GET("/getPost/:id", postHandler.GetPostById)
	postGroup.PUT("/update", postHandler.UpdatePost)
	postGroup.POST("/uploadImage", postHandler.UploadPostImage)
	//postGroup.DELETE("/delete/:id", postHandler.DeletePostById)
	//
	// login
	authHandler := handlers.NewLoginAndRegisterHandler()
	authGroup := e.Group("/auth")
	authGroup.GET("/", authHandler.Home)
	authGroup.GET("/:provider", authHandler.SignInWithProvider)
	authGroup.GET("/:provider/callback", authHandler.CallbackHandler)
	authGroup.GET("/success", authHandler.Success)

	// users group
	//userGroup := e.Group("/user")
	//userGroup.GET("")

	err := e.Start("localhost:" + os.Getenv("SERVER_PORT"))
	if err != nil {
		panic(err)
	}
}

func startDatabase() {
	config.Init()
}
