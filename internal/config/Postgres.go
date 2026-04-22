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
	)

	return db
}
