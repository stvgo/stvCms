package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"stvCms/internal/models"

	"github.com/redis/go-redis/v9"
)

type INotificationRepository interface {
	Save(ctx context.Context, notification models.Notification) error
	GetAll(ctx context.Context) ([]models.Notification, error)
	MarkRead(ctx context.Context, id string) error
	MarkAllRead(ctx context.Context) error
	GetUnreadCount(ctx context.Context) (int64, error)
	Delete(ctx context.Context, id string) error
}

type notificationRepository struct {
	redis *redis.Client
}

func NewNotificationRepository(redisClient *redis.Client) *notificationRepository {
	return &notificationRepository{redis: redisClient}
}

const notificationKey = "admin:notifications"

func (nr *notificationRepository) Save(ctx context.Context, n models.Notification) error {
	data, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %w", err)
	}
	return nr.redis.RPush(ctx, notificationKey, data).Err()
}

func (nr *notificationRepository) GetAll(ctx context.Context) ([]models.Notification, error) {
	data, err := nr.redis.LRange(ctx, notificationKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	notifications := make([]models.Notification, 0, len(data))
	for _, item := range data {
		var n models.Notification
		if err := json.Unmarshal([]byte(item), &n); err != nil {
			continue
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (nr *notificationRepository) MarkRead(ctx context.Context, id string) error {
	data, err := nr.redis.LRange(ctx, notificationKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for i, item := range data {
		var n models.Notification
		if err := json.Unmarshal([]byte(item), &n); err != nil {
			continue
		}
		if n.ID == id {
			n.Read = true
			updated, err := json.Marshal(n)
			if err != nil {
				return err
			}
			return nr.redis.LSet(ctx, notificationKey, int64(i), updated).Err()
		}
	}
	return nil
}

func (nr *notificationRepository) MarkAllRead(ctx context.Context) error {
	data, err := nr.redis.LRange(ctx, notificationKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for i, item := range data {
		var n models.Notification
		if err := json.Unmarshal([]byte(item), &n); err != nil {
			continue
		}
		if !n.Read {
			n.Read = true
			updated, err := json.Marshal(n)
			if err != nil {
				return err
			}
			if err := nr.redis.LSet(ctx, notificationKey, int64(i), updated).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (nr *notificationRepository) GetUnreadCount(ctx context.Context) (int64, error) {
	data, err := nr.redis.LRange(ctx, notificationKey, 0, -1).Result()
	if err != nil {
		return 0, err
	}

	var count int64
	for _, item := range data {
		var n models.Notification
		if err := json.Unmarshal([]byte(item), &n); err != nil {
			continue
		}
		if !n.Read {
			count++
		}
	}
	return count, nil
}

func (nr *notificationRepository) Delete(ctx context.Context, id string) error {
	data, err := nr.redis.LRange(ctx, notificationKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, item := range data {
		var n models.Notification
		if err := json.Unmarshal([]byte(item), &n); err != nil {
			continue
		}
		if n.ID == id {
			// Remove by value
			return nr.redis.LRem(ctx, notificationKey, 1, item).Err()
		}
	}
	return nil
}

// CleanupOldNotifications removes notifications older than the given duration
func (nr *notificationRepository) CleanupOldNotifications(ctx context.Context, olderThan time.Duration) error {
	data, err := nr.redis.LRange(ctx, notificationKey, 0, -1).Result()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-olderThan)
	for _, item := range data {
		var n models.Notification
		if err := json.Unmarshal([]byte(item), &n); err != nil {
			continue
		}
		if n.CreatedAt.Before(cutoff) {
			nr.redis.LRem(ctx, notificationKey, 1, item)
		}
	}
	return nil
}