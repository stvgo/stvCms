package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"stvCms/internal/models"
	"stvCms/internal/repository"
	"stvCms/internal/rest/request"
	"stvCms/internal/rest/response"
	"time"

	"stvCms/internal/clients"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type IPostService interface {
	CreatePost(req request.CreatePostRequest) (string, error)
	GetPosts() ([]response.PostResponse, error)
	GetPostById(id int) (response.PostResponse, error)
	GetPostByFilter(filter string) ([]response.PostResponse, error)
	UpdatePost(req request.UpdatePostRequest) (string, error)
	DeletePostById(id string) (string, error)
	SaveImage(image multipart.File, handler *multipart.FileHeader) (string, error)
	GetImage(filename string) ([]byte, error)
	AutoCompleteAI(reqAI request.AI) (string, error)
}

type postService struct {
	repository       repository.IPostRepository
	ctx              context.Context
	redisClient      *redis.Client
	openRouterClient clients.IOpenRouterClient
	r2               *s3.Client
}

func (ps *postService) GetImage(filename string) ([]byte, error) {
	img, err := ps.r2.GetObject(ps.ctx, &s3.GetObjectInput{
		Bucket: aws.String("stv-cms"),
		Key:    aws.String(filename),
	})
	if err != nil {
		return nil, err
	}
	defer img.Body.Close()

	data, err := io.ReadAll(img.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func NewPostService(ctx context.Context, redis redis.Client, openRouterClient clients.IOpenRouterClient, db *gorm.DB, r2 *s3.Client) *postService {
	return &postService{
		repository:       repository.NewPostGormRepository(db),
		ctx:              ctx,
		redisClient:      &redis,
		openRouterClient: openRouterClient,
		r2:               r2,
	}
}

func (ps *postService) CreatePost(req request.CreatePostRequest) (string, error) {
	post := models.Post{
		Title:     req.Title,
		UserID:    req.UserID,
		IsVisible: req.IsVisible,
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

	_ = ps.redisClient.Del(ps.ctx, "posts:all").Err()

	return modelPost, nil
}

func (ps *postService) GetPosts() ([]response.PostResponse, error) {
	posts := []response.PostResponse{}
	var modelPosts []models.Post

	val, err := ps.redisClient.Get(ps.ctx, "posts:all").Result()
	if err == nil && val != "" {
		if err := json.Unmarshal([]byte(val), &modelPosts); err != nil {
			return posts, nil
		}
	} else {
		modelPosts, err = ps.repository.GetPosts()
		if err != nil {
			return posts, err
		}
		if len(modelPosts) > 0 {
			data, _ := json.Marshal(modelPosts)
			_ = ps.redisClient.Set(ps.ctx, "posts:all", string(data), 24*time.Hour).Err()
		}
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
			IsVisible:     post.IsVisible,
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
		IsVisible:     post.IsVisible,
		ContentBlocks: contentBlocks,
	}

	return postResponse, nil
}

func (ps *postService) GetPostByFilter(filter string) ([]response.PostResponse, error) {
	modelPosts, err := ps.repository.GetPostsByFilter(filter)
	if err != nil {
		return []response.PostResponse{}, err
	}

	var posts []response.PostResponse

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
			IsVisible:     post.IsVisible,
			ContentBlocks: contentBlocks,
		}
		posts = append(posts, data)
	}

	return posts, nil
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

	_ = ps.redisClient.Del(ps.ctx, "posts:all").Err()

	return result, nil
}

func (ps *postService) DeletePostById(id string) (string, error) {
	postId, _ := strconv.Atoi(id)

	ok := ps.repository.DeletePostById(postId)

	if !ok {
		return "", fmt.Errorf("Error al borrar el post")
	}

	_ = ps.redisClient.Del(ps.ctx, "posts:all").Err()

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
			slog.Info("imagen redimensionada", "original",
				fmt.Sprintf("%dx%d", bounds.Dx(), bounds.Dy()), "new",
				fmt.Sprintf("%dx%d", resizedImg.Bounds().Dx(), resizedImg.Bounds().Dy()))
		}
	}

	ext := filepath.Ext(handler.Filename)
	if ext == "" {
		ext = "." + format
	}
	fileName := uuid.New().String() + ext

	imgR2, _, _ := imageToReader(format, resizedImg)
	return ps.uploadImageR2("stv-cms", fileName, imgR2)
}
func imageToReader(format string, img image.Image) (io.Reader, int64, error) {
	var buf bytes.Buffer

	if format == "jpeg" {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, 0, err
		}
		return &buf, int64(buf.Len()), nil
	} else if format == "png" {
		if err := png.Encode(&buf, img); err != nil {
			return nil, 0, err
		}
		return &buf, int64(buf.Len()), nil
	}

	return nil, 0, fmt.Errorf("format %s not supported", format)
}

func (ps *postService) uploadImageR2(bucket, filename string, body io.Reader) (string, error) {
	_, err := ps.r2.PutObject(ps.ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   body,
	})

	if err != nil {
		return errors.New("error al subir imagen a R2").Error(), err
	}
	return filename, err
}

func (ps *postService) AutoCompleteAI(reqAI request.AI) (string, error) {
	if reqAI.TextAI != "" {
		formatPrompt := "Realiza un formateo a este texto pero separalo por parrafos y que quede con formato html"
		prePrompt := fmt.Sprintf("Complementa el siguiente texto para crear un post de blog atractivo y bien estructurado."+
			" El texto debe ser claro, conciso y fácil de entender."+
			" Asegúrate de incluir una introducción, un desarrollo y una conclusión. El tema del post es: %s. %s", reqAI.TextAI, formatPrompt)

		text, err := ps.openRouterClient.GenAI(prePrompt)
		if err != nil {
			slog.Error("error al generar el texto", "error", err)
			return "", err
		}

		return text, nil
	}
	slog.Info("text_ai is empty")

	if reqAI.CodeAI != "" {
		formatCode := "Genera solo codigo en tipo main.go, no incluyas comentarios"
		prompt := fmt.Sprintf("Genera el siguiente código con buenas prácticas, moderno y estructurado: %s. %s", reqAI.CodeAI, formatCode)

		code, err := ps.openRouterClient.GenAI(prompt)
		if err != nil {
			slog.Error("error al generar el texto", "error", err)
		}
		return code, nil
	}
	slog.Info("code_ai is empty")

	return fmt.Sprintf("No se puede autocompletar AI para Text=null, Code=null "), nil

}
