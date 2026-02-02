package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	Title         string
	UserID        string
	ContentBlocks []ContentBlock
}

type ContentBlock struct {
	gorm.Model
	Type     string `gorm:"type:varchar(20)"` // "text", "code", "image", "url"
	Order    int
	Content  string `gorm:"type:text"`
	Language string `gorm:"type:varchar(50)"`

	//
	PostID uint
	Post   Post `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
