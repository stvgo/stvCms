package handlers

import (
	"strconv"
	"stvCms/internal/rest/request"
	"stvCms/internal/rest/response"
	"stvCms/internal/services"

	"github.com/go-fuego/fuego"
)

type postHandler struct {
	service services.IPostService
}

func NewPostHandler() *postHandler {
	return &postHandler{
		service: services.NewPostService(),
	}
}

func (h *postHandler) CreatePost(ctx fuego.ContextWithBody[request.CreatePostRequest]) (string, error) {
	input, err := ctx.Body()
	if err != nil {
		return "", err
	}

	response, err := h.service.CreatePost(input)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (h *postHandler) GetPosts(ctx fuego.ContextWithBody[response.PostResponse]) (*[]response.PostResponse, error) {
	_, err := ctx.Body()
	if err != nil {
		return nil, err
	}

	responsePosts, err := h.service.GetPosts()
	if err != nil {
		return nil, err
	}

	return &responsePosts, nil
}

func (h *postHandler) UpdatePost(ctx fuego.ContextWithBody[request.UpdatePostRequest]) (string, error) {
	req, err := ctx.Body()

	response, err := h.service.UpdatePost(req)

	if err != nil {
		return "", err
	}

	return response, nil
}

func (h *postHandler) DeletePostById(ctx fuego.ContextNoBody) (any, string) {
	id := ctx.PathParam("id")

	_, err := h.service.DeletePostById(id)
	if err != nil {
		return nil, err.Error()
	}
	return nil, "Deleted"
}

func (h *postHandler) GetPostById(ctx fuego.ContextWithBody[response.PostResponse]) (*response.PostResponse, error) {
	id := ctx.PathParam("id")

	postId, _ := strconv.Atoi(id)
	responsePost, err := h.service.GetPostById(postId)

	if err != nil {
		return nil, err
	}

	return &responsePost, nil
}
