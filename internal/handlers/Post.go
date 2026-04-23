package handlers

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"stvCms/internal/clients"
	"stvCms/internal/rest/request"
	"stvCms/internal/services"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type postHandler struct {
	service services.IPostService
}

func NewPostHandler(ctx context.Context, redis redis.Client, openRouterClient clients.IOpenRouterClient, db *gorm.DB, r2 *s3.Client) *postHandler {
	return &postHandler{
		service: services.NewPostService(
			ctx,
			redis,
			openRouterClient,
			db,
			r2,
		),
	}
}

func (h *postHandler) CreatePost(c echo.Context) error {
	var input request.CreatePostRequest
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	result, err := h.service.CreatePost(input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *postHandler) GetPosts(c echo.Context) error {
	responsePosts, err := h.service.GetPosts()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, responsePosts)
}

func (h *postHandler) UpdatePost(c echo.Context) error {
	var req request.UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	result, err := h.service.UpdatePost(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *postHandler) DeletePostById(c echo.Context) error {
	id := c.Param("id")

	_, err := h.service.DeletePostById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusNoContent, echo.Map{"message": "Deleted"})
}

func (h *postHandler) GetPostById(c echo.Context) error {
	id := c.Param("id")

	postId, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}

	responsePost, err := h.service.GetPostById(postId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, responsePost)
}

func (h *postHandler) UploadPostImage(c echo.Context) error {

	file, handler, err := c.Request().FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Failed to get image from request"})
	}
	defer file.Close()

	filename, err := h.service.SaveImage(file, handler)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message":  "Image uploaded successfully",
		"filename": filename,
	})
}

func (h *postHandler) GetImage(c echo.Context) error {
	filename := c.Param("filename")

	image, err := h.service.GetImage(filename)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	c.Response().Header().Set("Content-Type", "image/jpeg")
	return c.Stream(http.StatusOK, "image/jpeg", bytes.NewReader(image))
}

func (h *postHandler) GetPostByFilter(c echo.Context) error {
	filter := c.Param("filter")

	if filter == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "filter is empty"})
	}

	responsePost, err := h.service.GetPostByFilter(filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err})
	}

	return c.JSON(http.StatusOK, responsePost)
}

func (h *postHandler) AutoCompleteAI(c echo.Context) error {
	reqAI := request.AI{}
	err := c.Bind(&reqAI)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "text_ai is empty"})
	}

	response, err := h.service.AutoCompleteAI(reqAI)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err})
	}

	return c.JSON(http.StatusOK, response)

}
