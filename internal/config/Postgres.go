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
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_HOST := os.Getenv("DB_HOST")
	DB_PORT := os.Getenv("DB_PORT")
	DB := os.Getenv("DB")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB)

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
