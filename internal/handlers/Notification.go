package handlers

import (
	"net/http"

	"stvCms/internal/repository"
	"stvCms/internal/services"

	"github.com/labstack/echo/v4"
)

type notificationHandler struct {
	service services.INotificationService
}

func NewNotificationHandler(notifRepo repository.INotificationRepository) *notificationHandler {
	svc := services.NewNotificationService(notifRepo)
	return &notificationHandler{service: svc}
}

func (h *notificationHandler) GetAll(c echo.Context) error {
	notifications, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, notifications)
}

func (h *notificationHandler) GetUnreadCount(c echo.Context) error {
	count, err := h.service.GetUnreadCount(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"unread_count": count})
}

func (h *notificationHandler) MarkRead(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "id is required"})
	}
	if err := h.service.MarkRead(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "notification marked as read"})
}

func (h *notificationHandler) MarkAllRead(c echo.Context) error {
	if err := h.service.MarkAllRead(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "all notifications marked as read"})
}

func (h *notificationHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "id is required"})
	}
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "notification deleted"})
}