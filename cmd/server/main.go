package main

import (
	"net/http"
	"os"
	"stvCms/internal/config"
	"stvCms/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	startDatabase()
	startServer()
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
}

func startServer() {
	router := gin.Default()
	router.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"response": "pong!"})
	})
	// post group

	postHandler := handlers.NewPostHandler()
	login := handlers.NewLoginAndRegisterHandler()

	postGroup := router.Group("/post")
	postGroup.POST("/create", postHandler.CreatePost)
	postGroup.GET("/getAll", postHandler.GetPosts)
	postGroup.GET("/getPost/:id", postHandler.GetPostById)
	postGroup.PUT("/update", postHandler.UpdatePost)
	postGroup.DELETE("/delete/:id", postHandler.DeletePostById)
	postGroup.POST("/codeContent", postHandler.InsertCodeContentInPost)
	postGroup.POST("/:id/image", postHandler.UpdloadImage)

	// login
	postGroup.POST("/login/oauth2", login.Login)

	// users group
	//userGroup := router.Group("/user")
	//userGroup.GET("")

	err := router.Run(":" + os.Getenv("GIN_PORT"))
	if err != nil {
		return
	}
}

func startDatabase() {
	config.Init()
}
