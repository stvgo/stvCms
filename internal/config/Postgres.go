package config

import (
	"fmt"
	"log"
	"os"
	"stvCms/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		DB_USER := os.Getenv("DB_USER")
		DB_PASSWORD := os.Getenv("DB_PASSWORD")
		DB_HOST := os.Getenv("DB_HOST")
		DB_PORT := os.Getenv("DB_PORT")
		DB := os.Getenv("DB")
		log.Printf("DB config — USER=%q HOST=%q PORT=%q DB=%q", DB_USER, DB_HOST, DB_PORT, DB)
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB)
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
