package main

import (
	"context"
	"log"
	"stvCms/internal/config"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"gorm.io/gorm"
)

func main() {
	loadEnv()
	cfg := config.Load()
	initAuth(cfg)
	db := config.InitDB(cfg)
	startServer(cfg, db)
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from system environment")
	}
}

func initAuth(cfg *config.Config) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.ClientCallbackURL == "" {
		log.Println("Warning: Google OAuth not configured, skipping auth setup")
		return
	}

	goth.UseProviders(
		google.New(cfg.ClientID, cfg.ClientSecret, cfg.ClientCallbackURL),
	)

	gothic.Store = sessions.NewCookieStore([]byte(cfg.SessionSecret))
}

func startServer(cfg *config.Config, db *gorm.DB) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	ctx := context.Background()
	registerRoutes(e, cfg, db, ctx)

	if err := e.Start("0.0.0.0:" + cfg.Port); err != nil {
		panic(err)
	}
}
