package handlers

import (
	"net/http"
	"strconv"
	"stvCms/internal/rest/request"
	"stvCms/internal/services"

	"github.com/labstack/echo/v4"
)

type postHandler struct {
	service services.IPostService
}

func NewPostHandler() *postHandler {
	return &postHandler{
		service: services.NewPostService(),
	}
}

func (h *postHandler) CreatePost(c echo.Context) error {
	var input request.CreatePostRequest
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := h.service.CreatePost(input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *postHandler) GetPosts(c echo.Context) error {
	responsePosts, err := h.service.GetPosts()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, responsePosts)
}

func (h *postHandler) UpdatePost(c echo.Context) error {
	var req request.UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := h.service.UpdatePost(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *postHandler) DeletePostById(c echo.Context) error {
	id := c.Param("id")

	_, err := h.service.DeletePostById(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusNoContent, map[string]string{"message": "Deleted"})
}

func (h *postHandler) GetPostById(c echo.Context) error {
	id := c.Param("id")

	postId, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	responsePost, err := h.service.GetPostById(postId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, responsePost)
}

func (h *postHandler) UploadPostImage(c echo.Context) error {

	file, handler, err := c.Request().FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to get image from request"})
	}
	defer file.Close()

	filename, err := h.service.SaveImage(file, handler)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":  "Image uploaded successfully",
		"filename": filename,
	})
}
