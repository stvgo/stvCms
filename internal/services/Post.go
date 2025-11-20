package services

import (
	"fmt"
	"strconv"
	"stvCms/internal/models"
	"stvCms/internal/repository"
	"stvCms/internal/rest/request"
	"stvCms/internal/rest/response"

	"gorm.io/gorm"
)

type IPostService interface {
	CreatePost(req request.CreatePostRequest) (string, error)
	GetPosts() ([]response.PostResponse, error)
	GetPostById(id int) (response.PostResponse, error)
	UpdatePost(req request.UpdatePostRequest) (string, error)
	DeletePostById(id string) (string, error)
}

type postService struct {
	repository repository.IPostRepository
}

func NewPostService() *postService {
	return &postService{
		repository: repository.NewPostGormRepository(),
	}
}

func (ps *postService) CreatePost(req request.CreatePostRequest) (string, error) {
	post := models.Post{
		Title:  req.Title,
		UserID: req.UserID,
	}

	for _, block := range req.ContentBlocks {

		contentBlock := models.ContentBlock{
			Type:     block.Type,
			Order:    block.Order,
			Content:  block.Content,
			Language: block.Language,
		}
		post.ContentBlocks = append(post.ContentBlocks, contentBlock)
	}

	modelPost, err := ps.repository.CreatePost(post)
	if err != nil {
		return "No se pudo crear el post", err
	}

	return modelPost, nil
}

func (ps *postService) GetPosts() ([]response.PostResponse, error) {
	posts := []response.PostResponse{}
	modelPosts, err := ps.repository.GetPosts()
	if err != nil {
		return posts, nil
	}

	for _, post := range modelPosts {
		var contentBlocks []response.ContentBlockResponse
		for _, block := range post.ContentBlocks {
			contentBlocks = append(contentBlocks, response.ContentBlockResponse{
				Id:       block.ID,
				Type:     block.Type,
				Order:    block.Order,
				Content:  block.Content,
				Language: block.Language,
			})
		}

		data := response.PostResponse{
			Id:            post.Model.ID,
			CreatedAt:     post.CreatedAt,
			UpdatedAt:     post.UpdatedAt,
			Title:         post.Title,
			UserID:        post.UserID,
			ContentBlocks: contentBlocks,
		}
		posts = append(posts, data)
	}

	return posts, nil
}

func (ps *postService) GetPostById(id int) (response.PostResponse, error) {

	post, err := ps.repository.GetPostById(uint(id))
	if err != nil {
		return response.PostResponse{}, err
	}

	var contentBlocks []response.ContentBlockResponse
	for _, block := range post.ContentBlocks {
		contentBlocks = append(contentBlocks, response.ContentBlockResponse{
			Id:       block.ID,
			Type:     block.Type,
			Order:    block.Order,
			Content:  block.Content,
			Language: block.Language,
		})
	}

	postResponse := response.PostResponse{
		Id:            post.Model.ID,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
		Title:         post.Title,
		UserID:        post.UserID,
		ContentBlocks: contentBlocks,
	}

	return postResponse, nil
}

func (ps *postService) UpdatePost(req request.UpdatePostRequest) (string, error) {
	postModel, err := ps.repository.GetPostById(req.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "Post no encontrado con ese ID", err
		}
		return "Error al buscar el post", err
	}

	// Actualizar título si se proporciona
	if req.Title != "" {
		postModel.Title = req.Title
	}

	// Actualizar content blocks si se proporcionan
	if len(req.ContentBlocks) > 0 {
		// Limpiar los bloques existentes
		postModel.ContentBlocks = []models.ContentBlock{}

		// Agregar los nuevos bloques
		for _, block := range req.ContentBlocks {
			contentBlock := models.ContentBlock{
				Type:     block.Type,
				Order:    block.Order,
				Content:  block.Content,
				Language: block.Language,
			}
			postModel.ContentBlocks = append(postModel.ContentBlocks, contentBlock)
		}
	}

	postUpdated, err := ps.repository.UpdatePost(req.Id, postModel)
	if err != nil {
		return "", err
	}

	return postUpdated, nil
}

func (ps *postService) DeletePostById(id string) (string, error) {
	postId, _ := strconv.Atoi(id)

	ok := ps.repository.DeletePostById(postId)

	if !ok {
		return "", fmt.Errorf("Error al borrar el post")
	}

	return "Post borrado", nil
}
