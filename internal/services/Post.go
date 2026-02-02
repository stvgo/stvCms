package services

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"stvCms/internal/models"
	"stvCms/internal/repository"
	"stvCms/internal/rest/request"
	"stvCms/internal/rest/response"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IPostService interface {
	CreatePost(req request.CreatePostRequest) (string, error)
	GetPosts() ([]response.PostResponse, error)
	GetPostById(id int) (response.PostResponse, error)
	UpdatePost(req request.UpdatePostRequest) (string, error)
	DeletePostById(id string) (string, error)
	SaveImage(image multipart.File, handler *multipart.FileHeader) (string, error)
	GetImage(filename string) ([]byte, error)
}

type postService struct {
	repository repository.IPostRepository
}

func (ps *postService) GetImage(filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join("././public/uploads", filename))
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
	ok := ps.repository.ExistsPost(int(req.Id))
	if !ok {
		return "", gorm.ErrRecordNotFound
	}

	var contentBlocks []models.ContentBlock

	for _, block := range req.ContentBlocks {
		contentBlocks = append(contentBlocks, models.ContentBlock{
			Type:     block.Type,
			Order:    block.Order,
			Content:  block.Content,
			Language: block.Language,
		})
	}
	postModel := models.Post{
		Title:         req.Title,
		ContentBlocks: contentBlocks,
	}

	result, err := ps.repository.UpdatePost(req.Id, postModel)

	if err != nil {
		return "", err
	}

	return result, nil
}

func (ps *postService) DeletePostById(id string) (string, error) {
	postId, _ := strconv.Atoi(id)

	ok := ps.repository.DeletePostById(postId)

	if !ok {
		return "", fmt.Errorf("Error al borrar el post")
	}

	return "Post borrado", nil
}

func (ps *postService) SaveImage(imageFile multipart.File, handler *multipart.FileHeader) (string, error) {
	maxSize := int64(10 << 20) // 10 MB
	if handler.Size > maxSize {
		return "", fmt.Errorf("el archivo excede el tamaño máximo permitido de 10MB")
	}

	img, format, err := image.Decode(imageFile)
	if err != nil {
		return "", fmt.Errorf("error al decodificar la imagen: %w", err)
	}

	maxWidth := 1920
	maxHeight := 1920
	minSizeForResize := int64(3 << 20) // 3 MB

	resizedImg := img
	bounds := img.Bounds()

	if (handler.Size > minSizeForResize) || (bounds.Dx() > maxWidth || bounds.Dy() > maxHeight) {
		if bounds.Dx() > maxWidth || bounds.Dy() > maxHeight {
			resizedImg = imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)
			slog.Info("imagen redimensionada", "original", fmt.Sprintf("%dx%d", bounds.Dx(), bounds.Dy()), "new", fmt.Sprintf("%dx%d", resizedImg.Bounds().Dx(), resizedImg.Bounds().Dy()))
		}
	}

	ext := filepath.Ext(handler.Filename)
	if ext == "" {
		ext = "." + format
	}
	fileName := uuid.New().String() + ext

	uploadDir := "././public/uploads"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		slog.Error("error al crear directorio", "error", err)
		return "", fmt.Errorf("error al crear directorio: %w", err)
	}

	outputPath := filepath.Join(uploadDir, fileName)
	err = imaging.Save(resizedImg, outputPath)
	if err != nil {
		slog.Error("error al guardar imagen", "error", err)
		return "", fmt.Errorf("error al guardar imagen: %w", err)
	}

	slog.Info("imagen guardada correctamente", "filename", fileName, "width", resizedImg.Bounds().Dx(), "height", resizedImg.Bounds().Dy())

	return fileName, nil
}
