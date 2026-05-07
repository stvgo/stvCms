package config

import (
	"log/slog"
	"os"
)

type Config struct {
	// Server
	Port          string
	SessionSecret string
	AuthSecret    string
	// OAuth - Google
	ClientID          string
	ClientSecret      string
	ClientCallbackURL string
	// OAuth - GitHub
	GitHubClientID          string
	GitHubClientSecret      string
	GitHubClientCallbackURL string
	// Database
	DatabaseURL string
	DBUser      string
	DBPassword  string
	DBHost      string
	DBPort      string
	DBName      string
	// Redis
	RedisURL      string
	RedisAddr     string
	RedisPassword string
	// AI
	OpenRouterAPIKey string
	// Cloudflare R2
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
}

func Load() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "default-secret-key-change-this"
		slog.Warn("Warning: SESSION_SECRET not set, using default key")
	}

	authSecret := os.Getenv("AUTH_SECRET")
	if authSecret == "" {
		authSecret = "default-auth-secret-key-change-this-min-32"
		slog.Warn("Warning: AUTH_SECRET not set, using default key")
	}

	return &Config{
		Port:              port,
		SessionSecret:     sessionSecret,
		AuthSecret:        authSecret,
		ClientID:          os.Getenv("CLIENT_ID"),
		ClientSecret:      os.Getenv("CLIENT_SECRET"),
		ClientCallbackURL: os.Getenv("CLIENT_CALLBACK_URL"),
		GitHubClientID:          os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret:      os.Getenv("GITHUB_CLIENT_SECRET"),
		GitHubClientCallbackURL: os.Getenv("GITHUB_CALLBACK_URL"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		DBUser:            os.Getenv("DB_USER"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBHost:            os.Getenv("DB_HOST"),
		DBPort:            os.Getenv("DB_PORT"),
		DBName:            os.Getenv("DB"),
		RedisURL:          os.Getenv("REDIS_URL"),
		RedisAddr:         os.Getenv("REDIS_ADDR"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		OpenRouterAPIKey:  os.Getenv("OPEN_ROUTER_API_KEY"),
		AccountID:         os.Getenv("ACCOUNT_ID"),
		AccessKeyID:       os.Getenv("ACCESS_KEY_ID"),
		SecretAccessKey:   os.Getenv("SECRET_ACCESS_KEY"),
	}
}
