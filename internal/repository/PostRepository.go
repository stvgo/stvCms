package repository

import (
	"errors"
	"fmt"
	"stvCms/internal/models"

	"gorm.io/gorm"
)

type IPostRepository interface {
	CreatePost(post models.Post) (string, error)
	GetPosts(userID string) ([]models.Post, error)
	GetPublicPosts() ([]models.Post, error)
	GetPendingPosts() ([]models.Post, error)
	UpdatePost(id uint, post models.Post) (string, error)
	GetPostById(id uint, userID string) (models.Post, error)
	GetPublicPostById(id uint) (models.Post, error)
	GetPostsByFilter(filter string, userID string) ([]models.Post, error)
	DeletePostById(id int) bool
	ExistsPost(id int) bool
}

type postRepository struct {
	db *gorm.DB
}

func (pr *postRepository) GetPostsByFilter(filter string, userID string) ([]models.Post, error) {
	var posts []models.Post
	err := pr.db.Preload("ContentBlocks").
		Where("status = ? OR user_id = ?", "public", userID).
		Where("(title LIKE ? OR user_id LIKE ? OR id IN (SELECT post_id FROM content_blocks WHERE content LIKE ?))",
			"%"+filter+"%", "%"+filter+"%", "%"+filter+"%").
		Find(&posts).Error
	if err != nil {
		return posts, err
	}
	return posts, nil
}

func NewPostGormRepository(db *gorm.DB) *postRepository {
	return &postRepository{db: db}
}

func (pr *postRepository) GetPendingPosts() ([]models.Post, error) {
	var posts []models.Post
	err := pr.db.Preload("ContentBlocks").
		Where("status = ?", "pending").
		Find(&posts).Error
	if err != nil {
		return posts, err
	}
	return posts, nil
}

func (pr *postRepository) CreatePost(post models.Post) (string, error) {
	err := pr.db.Create(&post).Error
	if err != nil {
		return "No se pudo crear el post", err
	}
	return fmt.Sprintf("%d", post.ID), nil
}

func (pr *postRepository) GetPosts(userID string) ([]models.Post, error) {
	var posts []models.Post
	err := pr.db.Preload("ContentBlocks").
		Where("status = ? OR user_id = ?", "public", userID).
		Find(&posts).Error
	if err != nil {
		return posts, err
	}
	return posts, nil
}

func (pr *postRepository) GetPublicPosts() ([]models.Post, error) {
	var posts []models.Post
	err := pr.db.Preload("ContentBlocks").
		Where("status = ?", "public").
		Find(&posts).Error
	if err != nil {
		return posts, err
	}
	return posts, nil
}

func (pr *postRepository) UpdatePost(id uint, post models.Post) (string, error) {
	result := pr.db.Model(&models.Post{}).Where("id = ?", id).Updates(post)
	if result.Error != nil {
		return "No se pudo actualizar el post", result.Error
	}

	if result.RowsAffected == 0 {
		return "El post no fue encontrado o no habían datos para actualizar", nil
	}

	return "Post actualizado", nil
}

func (pr *postRepository) GetPostById(id uint, userID string) (models.Post, error) {
	var post models.Post
	if err := pr.db.Where("id = ? AND (status = ? OR user_id = ?)", id, "public", userID).First(&post).Error; err != nil {
		return post, err
	}

	contentBlocks, err := pr.GetContentBlocksById(id)
	post.ContentBlocks = contentBlocks
	return post, err
}

func (pr *postRepository) GetPublicPostById(id uint) (models.Post, error) {
	var post models.Post
	if err := pr.db.Where("id = ? AND status = ?", id, "public").First(&post).Error; err != nil {
		return post, err
	}

	contentBlocks, err := pr.GetContentBlocksById(id)
	post.ContentBlocks = contentBlocks
	return post, err
}

func (pr *postRepository) GetContentBlocksById(id uint) ([]models.ContentBlock, error) {
	var contentBlocks []models.ContentBlock
	err := pr.db.Where("post_id = ?", id).Find(&contentBlocks).Error
	return contentBlocks, err
}

func (pr *postRepository) DeletePostById(id int) bool {
	ok := pr.db.Delete(&models.Post{}, id).RowsAffected > 0
	return ok
}

func (pr *postRepository) ExistsPost(id int) bool {
	err := pr.db.First(&models.Post{}, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
