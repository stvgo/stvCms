package handlers

import (
	"net/http"
	"stvCms/internal/services"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	service services.IAuthService
}

func NewAuthHandler(service services.IAuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type SyncUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	GoogleID string `json:"google_id"`
}

func (h *AuthHandler) SyncUser(c echo.Context) error {
	var req SyncUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "email is required"})
	}

	user, err := h.service.SyncUser(req.Email, req.Name, req.Image, req.GoogleID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Internal server error"})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Me(c echo.Context) error {
	email, ok := c.Get("userEmail").(string)
	if !ok || email == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	user, err := h.service.GetUserByEmail(email)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"image": user.Image,
		"role":  user.Role,
	})
}
