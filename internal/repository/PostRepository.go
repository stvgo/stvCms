package repository

import (
	"errors"
	"stvCms/internal/config"
	"stvCms/internal/models"

	"gorm.io/gorm"
)

type IPostRepository interface {
	CreatePost(post models.Post) (string, error)
	GetPosts() ([]models.Post, error)
	UpdatePost(id uint, post models.Post) (string, error)
	GetPostById(id uint) (models.Post, error)
	DeletePostById(id int) bool
	SavePostImage(postID int, publicUrl string) error
	GetPostImage(postID uint) (string, error)
}

type postRepository struct {
	db *gorm.DB
}

func NewPostGormRepository() *postRepository {
	return &postRepository{
		db: config.Init(),
	}
}

func (pr *postRepository) CreatePost(post models.Post) (string, error) {
	err := pr.db.Create(&post).Error
	if err != nil {
		return "No se pudo crear el post", err
	}
	return "Post creado", nil
}

func (pr *postRepository) GetPosts() ([]models.Post, error) {
	var posts []models.Post
	err := pr.db.Find(&posts).Error
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

func (pr *postRepository) GetPostById(id uint) (models.Post, error) {
	var post models.Post
	err := pr.db.First(&post, id).Error

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

func (pr *postRepository) SavePostImage(postID int, publicUrl string) error {
	// TODO
	return errors.New("Not implemented")
}

func (pr *postRepository) GetPostImage(postID uint) (string, error) {
	//TODO implement me
	panic("implement me")
}
