package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"stvCms/internal/rest/request"
	"stvCms/internal/services"
)

type postHandler struct {
	service services.IPostService
}

func NewPostHandler() *postHandler {
	return &postHandler{
		service: services.NewPostService(),
	}
}

func (h *postHandler) CreatePost(ctx *gin.Context) {
	postRequest := request.CreatePostRequest{}

	if err := ctx.ShouldBindJSON(&postRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.CreatePost(postRequest)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": response})

}

func (h *postHandler) GetPosts(ctx *gin.Context) {
	response, err := h.service.GetPosts()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

func (h *postHandler) UpdatePost(ctx *gin.Context) {
	req := request.UpdatePostRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.UpdatePost(req)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"response": response})

}

func (h *postHandler) DeletePostById(ctx *gin.Context) {
	id := ctx.Param("id")

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	response, err := h.service.DeletePostById(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

func (h *postHandler) GetPostById(ctx *gin.Context) {
	id := ctx.Param("id")

	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	response, err := h.service.GetPostById(id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *postHandler) InsertCodeContentInPost(ctx *gin.Context) {
	var req request.CodeContent
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.InsertCodeContentInPost(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": response})
}

func (h *postHandler) UpdloadImage(ctx *gin.Context) {
	postID := ctx.Param("id")
	if postID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.SavePostImage(postID, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"response": response})
}
