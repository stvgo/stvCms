package services

import (
	"context"
	"fmt"
	"time"

	"stvCms/internal/models"
	"stvCms/internal/repository"

	"github.com/google/uuid"
)

type INotificationService interface {
	NotifyPendingPost(ctx context.Context, postID uint, title, authorID, authorName string) error
	GetAll(ctx context.Context) ([]models.Notification, error)
	GetUnreadCount(ctx context.Context) (int64, error)
	MarkRead(ctx context.Context, id string) error
	MarkAllRead(ctx context.Context) error
	Delete(ctx context.Context, id string) error
}

type notificationService struct {
	repo repository.INotificationRepository
}

func NewNotificationService(repo repository.INotificationRepository) *notificationService {
	return &notificationService{repo: repo}
}

func (ns *notificationService) NotifyPendingPost(ctx context.Context, postID uint, title, authorID, authorName string) error {
	n := models.Notification{
		ID:         uuid.New().String(),
		Type:       "post_pending",
		Title:      "Nuevo post pendiente de aprobación",
		Message:    fmt.Sprintf("%s creó el post \"%s\" y está esperando aprobación", authorName, title),
		PostID:     postID,
		AuthorID:   authorID,
		AuthorName: authorName,
		Read:       false,
		CreatedAt:  time.Now(),
	}
	return ns.repo.Save(ctx, n)
}

func (ns *notificationService) GetAll(ctx context.Context) ([]models.Notification, error) {
	return ns.repo.GetAll(ctx)
}

func (ns *notificationService) GetUnreadCount(ctx context.Context) (int64, error) {
	return ns.repo.GetUnreadCount(ctx)
}

func (ns *notificationService) MarkRead(ctx context.Context, id string) error {
	return ns.repo.MarkRead(ctx, id)
}

func (ns *notificationService) MarkAllRead(ctx context.Context) error {
	return ns.repo.MarkAllRead(ctx)
}

func (ns *notificationService) Delete(ctx context.Context, id string) error {
	return ns.repo.Delete(ctx, id)
}