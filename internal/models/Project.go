package models

import "gorm.io/gorm"

type Project struct {
	gorm.Model
	Title       string `gorm:"type:varchar(200)"`
	Description string `gorm:"type:text"`
	Type        string `gorm:"type:varchar(20)"` // "game", "web", "api", "tool", "library"
	URL         string `gorm:"type:text"`
	EmbedURL    string `gorm:"type:text"`
	ImageURL    string `gorm:"type:varchar(500)"`
	GitHubURL   string `gorm:"type:text;column:git_hub_url"`
	TechStack   string `gorm:"type:text"` // comma-separated: "Go,WASM,Ebiten"
	UserID      string `gorm:"type:varchar(100)"`
}