package main

import (
	"net/http"
	"os"
	"stvCms/internal/config"
	"stvCms/internal/handlers"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
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
	s := fuego.NewServer()
	// post group

	postHandler := handlers.NewPostHandler()
	//login := handlers.NewLoginAndRegisterHandler()

	postGroup := fuego.Group(s, "/post")
	fuego.Post(postGroup, "/create", postHandler.CreatePost, option.DefaultStatusCode(http.StatusCreated))
	fuego.Get(postGroup, "/getAll", postHandler.GetPosts)
	fuego.Get(postGroup, "/getPost/{id}", postHandler.GetPostById)
	fuego.Put(postGroup, "/update", postHandler.UpdatePost)
	//fuego.Delete(postGroup, "/delete/{id}", postHandler.DeletePostById, option.DefaultStatusCode(http.StatusNoContent))
	//
	//// login
	//postGroup.POST("/login/oauth2", login.Login)

	// users group
	//userGroup := router.Group("/user")
	//userGroup.GET("")

	s.Addr = "localhost:" + os.Getenv("SERVER_PORT")
	err := s.Run()
	if err != nil {
		panic(err)
	}
}

func startDatabase() {
	config.Init()
}
