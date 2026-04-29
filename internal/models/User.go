package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;not null"`
	Name     string
	Image    string
	GoogleID string `gorm:"uniqueIndex"`
	Role     string `gorm:"default:'user'"`
}
