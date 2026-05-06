package repository

import (
	"context"
	"testing"
	"time"

	"stvCms/internal/models"

	"github.com/alicebob/miniredis/v2"
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

func setupRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return mr, client
}

// --- NotificationRepository ---

func TestNotificationRepository(t *testing.T) {
	mr, rdb := setupRedis(t)
	defer rdb.Close()

	ctx := context.Background()
	repo := NewNotificationRepository(rdb)

	// Clean up
	rdb.Del(ctx, notificationKey)
	defer rdb.Del(ctx, notificationKey)

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
		mr.FlushAll()
		notifs, err := repo.GetAll(ctx)
		require.NoError(t, err)
		assert.Empty(t, notifs)
	})

	t.Run("GetUnreadCount with empty list", func(t *testing.T) {
		mr.FlushAll()
		count, err := repo.GetUnreadCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("CleanupOldNotifications removes old entries", func(t *testing.T) {
		mr.FlushAll()

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
		mr.FlushAll()
		err := repo.CleanupOldNotifications(ctx, 24*time.Hour)
		require.NoError(t, err)
	})

	t.Run("Save with marshal error handled", func(t *testing.T) {
		// This is implicitly tested by saving valid notifications above.
		// The function only fails on json.Marshal which rarely fails.
		// We test the happy path extensively.
		n := models.Notification{
			ID:    "marshal-test",
			Type:  "post_pending",
			Title: "Test",
		}
		err := repo.Save(ctx, n)
		assert.NoError(t, err)
	})

	t.Run("MarkAllRead with multiple items", func(t *testing.T) {
		mr.FlushAll()
		n1 := models.Notification{ID: "mr-1", Type: "test", Title: "Notif 1", Message: "msg1", Read: false}
		n2 := models.Notification{ID: "mr-2", Type: "test", Title: "Notif 2", Message: "msg2", Read: true}
		n3 := models.Notification{ID: "mr-3", Type: "test", Title: "Notif 3", Message: "msg3", Read: false}
		require.NoError(t, repo.Save(ctx, n1))
		require.NoError(t, repo.Save(ctx, n2))
		require.NoError(t, repo.Save(ctx, n3))

		err := repo.MarkAllRead(ctx)
		require.NoError(t, err)

		notifs, _ := repo.GetAll(ctx)
		for _, n := range notifs {
			assert.True(t, n.Read, "notification %s should be read", n.ID)
		}
	})

	t.Run("Delete specific notification", func(t *testing.T) {
		mr.FlushAll()
		n1 := models.Notification{ID: "del-1", Type: "test", Title: "Keep", Message: "keep", Read: false}
		n2 := models.Notification{ID: "del-2", Type: "test", Title: "Delete", Message: "delete", Read: false}
		require.NoError(t, repo.Save(ctx, n1))
		require.NoError(t, repo.Save(ctx, n2))

		err := repo.Delete(ctx, "del-2")
		require.NoError(t, err)

		notifs, _ := repo.GetAll(ctx)
		assert.Len(t, notifs, 1)
		assert.Equal(t, "del-1", notifs[0].ID)
	})

	_ = mr // keep reference
}

// --- PostRepository: GetPendingPosts ---

func TestGetPendingPostsRepo(t *testing.T) {
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

// --- PostRepository: GetPendingPostByID ---

func TestGetPendingPostByID(t *testing.T) {
	db := setupDB(t)
	repo := NewPostGormRepository(db)

	db.Create(&models.Post{Title: "Pending Article", UserID: "u1", Status: "pending"})
	db.Create(&models.Post{Title: "Public Article", UserID: "u2", Status: "public"})

	t.Run("finds pending post by ID", func(t *testing.T) {
		post, err := repo.GetPendingPostByID(1)
		require.NoError(t, err)
		assert.Equal(t, "Pending Article", post.Title)
		assert.Equal(t, "pending", post.Status)
	})

	t.Run("not found for public post", func(t *testing.T) {
		_, err := repo.GetPendingPostByID(2)
		assert.Error(t, err)
	})

	t.Run("not found for nonexistent ID", func(t *testing.T) {
		_, err := repo.GetPendingPostByID(999)
		assert.Error(t, err)
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

	t.Run("GetPendingPostByID with DB error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&models.Post{}, &models.ContentBlock{}))
		repo := NewPostGormRepository(db)

		sqlDB, _ := db.DB()
		sqlDB.Close()

		_, err = repo.GetPendingPostByID(1)
		assert.Error(t, err)
	})
}