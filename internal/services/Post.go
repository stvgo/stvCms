package services

import (
	"bytes"
	"context"
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
	"time"

	"stvCms/internal/clients"
	"stvCms/internal/models"
	"stvCms/internal/repository"
	"stvCms/internal/rest/request"
	"stvCms/internal/rest/response"
	"stvCms/internal/services/enums"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const AdminEmail = "jsvaleriano321@gmail.com"

type IPostService interface {
	CreatePost(req request.CreatePostRequest) (string, error)
	GetPosts(userID string) ([]response.PostResponse, error)
	GetPublicPosts() ([]response.PostResponse, error)
	GetPostByID(id int, userID string) (response.PostResponse, error)
	GetPublicPostByID(id int) (response.PostResponse, error)
	GetPostByFilter(filter string, userID string) ([]response.PostResponse, error)
	UpdatePost(req request.UpdatePostRequest, userEmail string) (string, error)
	DeletePostByID(id string) (string, error)
	SaveImage(image multipart.File, handler *multipart.FileHeader) (string, error)
	GetImage(filename string) ([]byte, error)
	AutoCompleteAI(reqAI request.AI) (string, error)
	GetPendingPosts() ([]response.PostResponse, error)
	GetPendingPostByID(id uint) (response.PostResponse, error)
	ApprovePost(id uint) (string, error)
	RejectPost(id uint) (string, error)
}

type postService struct {
	repository       repository.IPostRepository
	notifRepo        repository.INotificationRepository
	ctx              context.Context
	redisClient      clients.IRedisClient
	openRouterClient clients.IOpenRouterClient
	r2               clients.IR2Client
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

func NewPostService(ctx context.Context, redisClient clients.IRedisClient, openRouterClient clients.IOpenRouterClient, db *gorm.DB, r2 clients.IR2Client, notifRepo repository.INotificationRepository) *postService {
	return &postService{
		repository:       repository.NewPostGormRepository(db),
		notifRepo:        notifRepo,
		ctx:              ctx,
		redisClient:      redisClient,
		openRouterClient: openRouterClient,
		r2:               r2,
	}
}

func (ps *postService) CreatePost(req request.CreatePostRequest) (string, error) {
	status := req.Status
	if status == "" {
		status = enums.PostStatusPublic
	}

	// Los posts públicos de usuarios no-admin quedan como pending hasta aprobación
	if status == enums.PostStatusPublic && req.UserEmail != AdminEmail {
		status = enums.PostStatusPending
	}

	post := models.Post{
		Title:  req.Title,
		UserID: req.UserID,
		Status: status,
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

	if status == enums.PostStatusPending {
		// Send notification to admin
		var postID uint
		fmt.Sscanf(modelPost, "%d", &postID)

		notification := models.Notification{
			ID:         uuid.New().String(),
			Type:       "post_pending",
			Title:      "Nuevo post pendiente de aprobación",
			Message:    fmt.Sprintf("%s creó el post \"%s\" y está esperando aprobación", req.UserID, req.Title),
			PostID:     postID,
			AuthorID:   req.UserID,
			AuthorName: req.UserID,
			Read:       false,
			CreatedAt:  time.Now(),
		}
		if ps.notifRepo != nil {
			if notifErr := ps.notifRepo.Save(context.Background(), notification); notifErr != nil {
				slog.Error("error al guardar notificación de post pendiente", "error", notifErr)
			}
		}

		return "Post creado — está pendiente de aprobación del administrador", nil
	}

	return modelPost, nil
}

func (ps *postService) GetPosts(userID string) ([]response.PostResponse, error) {
	posts := []response.PostResponse{}
	modelPosts, err := ps.repository.GetPosts(userID)
	if err != nil {
		return posts, err
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
			Id:            post.ID,
			CreatedAt:     post.CreatedAt,
			UpdatedAt:     post.UpdatedAt,
			Title:         post.Title,
			UserID:        post.UserID,
			Status:        post.Status,
			ContentBlocks: contentBlocks,
		}
		posts = append(posts, data)
	}

	return posts, nil
}

func (ps *postService) GetPublicPosts() ([]response.PostResponse, error) {
	posts := []response.PostResponse{}
	modelPosts, err := ps.repository.GetPublicPosts()
	if err != nil {
		return posts, err
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
			Id:            post.ID,
			CreatedAt:     post.CreatedAt,
			UpdatedAt:     post.UpdatedAt,
			Title:         post.Title,
			UserID:        post.UserID,
			Status:        post.Status,
			ContentBlocks: contentBlocks,
		}
		posts = append(posts, data)
	}

	return posts, nil
}

func (ps *postService) GetPostByID(id int, userID string) (response.PostResponse, error) {
	post, err := ps.repository.GetPostById(uint(id), userID)
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
		Status:        post.Status,
		ContentBlocks: contentBlocks,
	}

	return postResponse, nil
}

func (ps *postService) GetPublicPostByID(id int) (response.PostResponse, error) {
	post, err := ps.repository.GetPublicPostById(uint(id))
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
		Status:        post.Status,
		ContentBlocks: contentBlocks,
	}

	return postResponse, nil
}

func (ps *postService) GetPostByFilter(filter string, userID string) ([]response.PostResponse, error) {
	modelPosts, err := ps.repository.GetPostsByFilter(filter, userID)
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
			Status:        post.Status,
			ContentBlocks: contentBlocks,
		}
		posts = append(posts, data)
	}

	return posts, nil
}

func (ps *postService) UpdatePost(req request.UpdatePostRequest, userEmail string) (string, error) {
	ok := ps.repository.ExistsPost(int(req.Id))
	if !ok {
		return "", gorm.ErrRecordNotFound
	}

	status := req.Status
	if status != "" {
		validStatus := status == enums.PostStatusPublic || status == enums.PostStatusPrivate || status == enums.PostStatusPending
		if !validStatus {
			return "", fmt.Errorf("status invalido: %s", status)
		}
		// Solo el admin puede cambiar status a public (aprobar)
		if status == enums.PostStatusPublic && userEmail != AdminEmail {
			return "", fmt.Errorf("no tienes permiso para publicar posts — un administrador debe aprobarlo")
		}
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

	if status != "" {
		postModel.Status = status
	}

	result, err := ps.repository.UpdatePost(req.Id, postModel)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (ps *postService) DeletePostByID(id string) (string, error) {
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

	return ps.uploadImageR2("stv-cms", fileName, imageToReader(format, resizedImg))
}

func imageToReader(format string, img image.Image) io.Reader {
	var buf bytes.Buffer

	if format == "jpeg" || format == "jpg" {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil
		}
		return &buf
	} else if format == "png" {
		if err := png.Encode(&buf, img); err != nil {
			return nil
		}
		return &buf
	}

	return nil
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
		systemPrompt := "Eres un escritor profesional de blogs. Generas contenido claro, conciso y atractivo." +
			" Usas HTML semántico (<h2>, <h3>, <p>, <strong>, <em>, <ul>, <ol>, <blockquote>)." +
			" Incluyes introducción, desarrollo y conclusión." +
			" No uses markdown, solo HTML válido dentro de un <div>." +
			" No incluyas explicaciones, solo el contenido HTML."

		userPrompt := fmt.Sprintf("Escribe un post de blog bien estructurado sobre: %s", reqAI.TextAI)

		text, err := ps.openRouterClient.GenAI(systemPrompt, userPrompt)
		if err != nil {
			slog.Error("error al generar el texto", "error", err)
			return "", err
		}

		return text, nil
	}
	slog.Info("text_ai is empty")

	if reqAI.CodeAI != "" {
		systemPrompt := "Eres un ingeniero de software experto. Generas código limpio, idiomático y bien estructurado." +
			" Solo devuelves código ejecutable, sin explicaciones ni comentarios innecesarios." +
			" Si el usuario especifica un lenguaje de programación, úsalo."

		userPrompt := fmt.Sprintf("Genera código para: %s", reqAI.CodeAI)

		code, err := ps.openRouterClient.GenAI(systemPrompt, userPrompt)
		if err != nil {
			slog.Error("error al generar el código", "error", err)
			return "", err
		}
		return code, nil
	}
	slog.Info("code_ai is empty")

	return "", fmt.Errorf("text_ai y code_ai están vacíos")
}

func (ps *postService) GetPendingPostByID(id uint) (response.PostResponse, error) {
	post, err := ps.repository.GetPendingPostByID(id)
	if err != nil {
		return response.PostResponse{}, err
	}

	contentBlocks := make([]response.ContentBlockResponse, 0, len(post.ContentBlocks))
	for _, block := range post.ContentBlocks {
		contentBlocks = append(contentBlocks, response.ContentBlockResponse{
			Id:       block.ID,
			Type:     block.Type,
			Order:    block.Order,
			Content:  block.Content,
			Language: block.Language,
		})
	}

	return response.PostResponse{
		Id:            post.Model.ID,
		CreatedAt:     post.CreatedAt,
		UpdatedAt:     post.UpdatedAt,
		Title:         post.Title,
		UserID:        post.UserID,
		Status:        post.Status,
		ContentBlocks: contentBlocks,
	}, nil
}

func (ps *postService) GetPendingPosts() ([]response.PostResponse, error) {
	modelPosts, err := ps.repository.GetPendingPosts()
	if err != nil {
		return nil, err
	}

	posts := make([]response.PostResponse, 0, len(modelPosts))
	for _, post := range modelPosts {
		contentBlocks := make([]response.ContentBlockResponse, 0, len(post.ContentBlocks))
		for _, block := range post.ContentBlocks {
			contentBlocks = append(contentBlocks, response.ContentBlockResponse{
				Id:       block.ID,
				Type:     block.Type,
				Order:    block.Order,
				Content:  block.Content,
				Language: block.Language,
			})
		}
		posts = append(posts, response.PostResponse{
			Id:            post.ID,
			CreatedAt:     post.CreatedAt,
			UpdatedAt:     post.UpdatedAt,
			Title:         post.Title,
			UserID:        post.UserID,
			Status:        post.Status,
			ContentBlocks: contentBlocks,
		})
	}
	return posts, nil
}

func (ps *postService) ApprovePost(id uint) (string, error) {
	ok := ps.repository.ExistsPost(int(id))
	if !ok {
		return "", gorm.ErrRecordNotFound
	}
	postModel := models.Post{
		Status: enums.PostStatusPublic,
	}
	result, err := ps.repository.UpdatePost(id, postModel)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (ps *postService) RejectPost(id uint) (string, error) {
	ok := ps.repository.ExistsPost(int(id))
	if !ok {
		return "", gorm.ErrRecordNotFound
	}
	okDel := ps.repository.DeletePostById(int(id))
	if !okDel {
		return "", fmt.Errorf("error al rechazar el post")
	}
	return "Post rechazado", nil
}
