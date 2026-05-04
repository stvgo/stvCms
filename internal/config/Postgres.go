package config

import (
	"fmt"
	"log"
	"stvCms/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg *Config) *gorm.DB {
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		log.Printf("DB config — USER=%q HOST=%q PORT=%q DB=%q", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	} else {
		log.Println("DB config — using DATABASE_URL")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
		panic("Error al cargar la bd")
	}

	db.AutoMigrate(
		&models.Post{},
		&models.ContentBlock{},
		&models.User{},
		&models.Project{},
	)

	db.Exec("UPDATE posts SET status = 'public' WHERE status = '' OR status IS NULL")

	// Auto-seed: insert default projects if table is empty
	var projectCount int64
	db.Model(&models.Project{}).Count(&projectCount)
	if projectCount == 0 {
		log.Println("Seed: inserting default projects...")
		db.Create(&models.Project{
			Title:       "Flappy Kiro",
			Description: "Clone de Flappy Bird construido con HTML5 Canvas puro — sin dependencias, sin build step. Pixel-art con temática Kiro (paleta purple/pink). Proyecto para la hackathon de AWS usando Kiro Code. ¡Juega directo en el navegador!",
			Type:        "game",
			URL:         "https://flappy-kiro-production.up.railway.app",
			EmbedURL:    "https://flappy-kiro-production.up.railway.app",
			GitHubURL:   "https://github.com/stvgo/flappy-kiro",
			TechStack:   "HTML5,Canvas,JavaScript",
			UserID:      "Stiven Valeriano",
		})
		log.Println("Seed: default projects inserted")
	}

	return db
}
