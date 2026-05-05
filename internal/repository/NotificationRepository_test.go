package repository

import (
	"context"
	"testing"
	"time"

	"stvCms/internal/models"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Post{}, &models.ContentBlock{}))
	return db
}

// --- NotificationRepository ---

func TestNotificationRepository(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping notification repository tests")
	}
	defer rdb.Close()

	repo := NewNotificationRepository(rdb)

	// Clean up before and after
	rdb.Del(ctx, "admin:notifications")
	defer rdb.Del(ctx, "admin:notifications")

	t.Run("Save and GetAll", func(t *testing.T) {
		n := models.Notification{
			ID:         "test-1",
			Type:       "post_pending",
			Title:      "Test Notification",
			Message:    "Someone did something",
			PostID:     1,
			AuthorID:   "user1",
			AuthorName: "Test User",
			Read:       false,
		}

		err := repo.Save(ctx, n)
		require.NoError(t, err)

		notifs, err := repo.GetAll(ctx)
		require.NoError(t, err)
		assert.Len(t, notifs, 1)
		assert.Equal(t, "test-1", notifs[0].ID)
		assert.Equal(t, "Test Notification", notifs[0].Title)
		assert.False(t, notifs[0].Read)
	})

	t.Run("GetUnreadCount", func(t *testing.T) {
		count, err := repo.GetUnreadCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("MarkRead", func(t *testing.T) {
		err := repo.MarkRead(ctx, "test-1")
		require.NoError(t, err)

		notifs, _ := repo.GetAll(ctx)
		require.Len(t, notifs, 1)
		assert.True(t, notifs[0].Read)

		count, _ := repo.GetUnreadCount(ctx)
		assert.Equal(t, int64(0), count)
	})

	t.Run("MarkRead nonexistent id does nothing", func(t *testing.T) {
		err := repo.MarkRead(ctx, "nonexistent-id")
		assert.NoError(t, err)
	})

	t.Run("MarkAllRead", func(t *testing.T) {
		n2 := models.Notification{
			ID:         "test-2",
			Type:       "post_pending",
			Title:      "Second",
			Message:    "Another event",
			PostID:     2,
			AuthorID:   "user2",
			AuthorName: "User Two",
			Read:       false,
		}
		repo.Save(ctx, n2)

		n3 := models.Notification{
			ID:         "test-3",
			Type:       "post_pending",
			Title:      "Third",
			Message:    "Third event",
			PostID:     3,
			AuthorID:   "user3",
			AuthorName: "Three",
			Read:       false,
		}
		repo.Save(ctx, n3)

		err := repo.MarkAllRead(ctx)
		require.NoError(t, err)

		count, _ := repo.GetUnreadCount(ctx)
		assert.Equal(t, int64(0), count)

		notifs, _ := repo.GetAll(ctx)
		for _, n := range notifs {
			assert.True(t, n.Read)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, "test-2")
		require.NoError(t, err)

		notifs, _ := repo.GetAll(ctx)
		for _, n := range notifs {
			assert.NotEqual(t, "test-2", n.ID)
		}
	})

	t.Run("Delete nonexistent id does nothing", func(t *testing.T) {
		err := repo.Delete(ctx, "nonexistent-id")
		assert.NoError(t, err)
	})

	t.Run("GetAll with empty list", func(t *testing.T) {
		rdb.Del(ctx, "admin:notifications")
		notifs, err := repo.GetAll(ctx)
		require.NoError(t, err)
		assert.Empty(t, notifs)
	})

	t.Run("GetUnreadCount with empty list", func(t *testing.T) {
		rdb.Del(ctx, "admin:notifications")
		count, err := repo.GetUnreadCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CleanupOldNotifications removes old entries", func(t *testing.T) {
		rdb.Del(ctx, "admin:notifications")

		oldNotif := models.Notification{
			ID:         "old-1",
			Type:       "post_pending",
			Title:      "Old Notif",
			Message:    "Old message",
			PostID:     1,
			AuthorID:   "u1",
			AuthorName: "Old User",
			Read:       false,
		}
		oldNotif.CreatedAt = time.Now().Add(-48 * time.Hour)
		repo.Save(ctx, oldNotif)

		recentNotif := models.Notification{
			ID:         "recent-1",
			Type:       "post_pending",
			Title:      "Recent",
			Message:    "Recent message",
			PostID:     2,
			AuthorID:   "u2",
			AuthorName: "Recent User",
			Read:       false,
			CreatedAt:  time.Now(),
		}
		repo.Save(ctx, recentNotif)

		err := repo.CleanupOldNotifications(ctx, 24*time.Hour)
		require.NoError(t, err)

		notifs, _ := repo.GetAll(ctx)
		assert.Len(t, notifs, 1)
		assert.Equal(t, "recent-1", notifs[0].ID)
	})

	t.Run("CleanupOldNotifications with no data", func(t *testing.T) {
		rdb.Del(ctx, "admin:notifications")
		err := repo.CleanupOldNotifications(ctx, 24*time.Hour)
		require.NoError(t, err)
	})
}

// --- PostRepository: GetPendingPosts ---

func TestGetPendingPosts(t *testing.T) {
	db := setupDB(t)
	repo := NewPostGormRepository(db)

	// Create posts with different statuses
	db.Create(&models.Post{Title: "Pending 1", UserID: "u1", Status: "pending"})
	db.Create(&models.Post{Title: "Public 1", UserID: "u2", Status: "public"})
	db.Create(&models.Post{Title: "Private 1", UserID: "u1", Status: "private"})
	db.Create(&models.Post{Title: "Pending 2", UserID: "u3", Status: "pending"})

	t.Run("returns only pending posts", func(t *testing.T) {
		pending, err := repo.GetPendingPosts()
		require.NoError(t, err)
		assert.Len(t, pending, 2)
		for _, p := range pending {
			assert.Equal(t, "pending", p.Status)
		}
	})

	t.Run("no pending posts", func(t *testing.T) {
		cleanDB := setupDB(t)
		cleanRepo := NewPostGormRepository(cleanDB)
		pending, err := cleanRepo.GetPendingPosts()
		require.NoError(t, err)
		assert.Empty(t, pending)
	})
}

// --- PostRepository: CreatePost returns ID ---

func TestCreatePost_ReturnsID(t *testing.T) {
	db := setupDB(t)
	repo := NewPostGormRepository(db)

	post := models.Post{
		Title:  "Test Post",
		UserID: "u1",
		Status: "public",
		ContentBlocks: []models.ContentBlock{
			{Type: "text", Order: 1, Content: "Hello"},
		},
	}

	result, err := repo.CreatePost(post)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	var count int64
	db.Model(&models.Post{}).Where("title = ?", "Test Post").Count(&count)
	assert.Equal(t, int64(1), count)
}

// --- PostRepository: Error paths ---

func TestPostRepository_ErrorPaths(t *testing.T) {
	t.Run("GetPendingPosts with DB error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&models.Post{}, &models.ContentBlock{}))
		repo := NewPostGormRepository(db)

		// Close the DB connection to force an error
		sqlDB, _ := db.DB()
		sqlDB.Close()

		_, err = repo.GetPendingPosts()
		assert.Error(t, err)
	})
}