package services

import (
	"context"
	"errors"
	"testing"

	"stvCms/internal/mocks"
	"stvCms/internal/models"
	"stvCms/internal/rest/request"
	"stvCms/internal/services/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Post{}, &models.ContentBlock{}, &models.User{}))
	return db
}

// --- LoginAndRegisterService ---

func TestNewLoginAndRegisterService(t *testing.T) {
	svc := NewLoginAndRegisterService()
	assert.NotNil(t, svc)
}

// --- CreatePost: pending approval for non-admin ---

func TestCreatePost_PendingForNonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIPostRepository(ctrl)
	svc := &postService{
		repository: repo,
		ctx:        context.Background(),
		notifRepo:  nil,
	}

	t.Run("non-admin public post becomes pending", func(t *testing.T) {
		repo.EXPECT().CreatePost(gomock.Any()).DoAndReturn(func(p models.Post) (string, error) {
			assert.Equal(t, enums.PostStatusPending, p.Status, "non-admin public post should be pending")
			assert.Equal(t, "regular@example.com", p.UserID)
			return "1", nil
		})

		req := request.CreatePostRequest{
			Title:     "My Post",
			UserID:    "regular@example.com",
			UserEmail: "regular@example.com",
			Status:    enums.PostStatusPublic,
		}
		result, err := svc.CreatePost(req)
		require.NoError(t, err)
		assert.Contains(t, result, "pendiente")
	})

	t.Run("admin public post stays public", func(t *testing.T) {
		repo.EXPECT().CreatePost(gomock.Any()).DoAndReturn(func(p models.Post) (string, error) {
			assert.Equal(t, enums.PostStatusPublic, p.Status, "admin post should stay public")
			return "2", nil
		})

		req := request.CreatePostRequest{
			Title:     "Admin Post",
			UserID:    "Stiven",
			UserEmail: "jsvaleriano321@gmail.com",
			Status:    enums.PostStatusPublic,
		}
		result, err := svc.CreatePost(req)
		require.NoError(t, err)
		assert.NotContains(t, result, "pendiente")
	})

	t.Run("private post stays private regardless of user", func(t *testing.T) {
		repo.EXPECT().CreatePost(gomock.Any()).DoAndReturn(func(p models.Post) (string, error) {
			assert.Equal(t, enums.PostStatusPrivate, p.Status)
			return "3", nil
		})

		req := request.CreatePostRequest{
			Title:     "Draft",
			UserID:    "regular@example.com",
			UserEmail: "regular@example.com",
			Status:    enums.PostStatusPrivate,
		}
		_, err := svc.CreatePost(req)
		require.NoError(t, err)
	})

	t.Run("default status becomes public for admin", func(t *testing.T) {
		repo.EXPECT().CreatePost(gomock.Any()).DoAndReturn(func(p models.Post) (string, error) {
			assert.Equal(t, enums.PostStatusPublic, p.Status)
			return "4", nil
		})

		req := request.CreatePostRequest{
			Title:     "No Status",
			UserID:    "Stiven",
			UserEmail: "jsvaleriano321@gmail.com",
		}
		_, err := svc.CreatePost(req)
		require.NoError(t, err)
	})

	t.Run("default status becomes pending for non-admin", func(t *testing.T) {
		repo.EXPECT().CreatePost(gomock.Any()).DoAndReturn(func(p models.Post) (string, error) {
			assert.Equal(t, enums.PostStatusPending, p.Status, "default for non-admin should be pending")
			return "5", nil
		})

		req := request.CreatePostRequest{
			Title:     "No Status Non-Admin",
			UserID:    "regular@example.com",
			UserEmail: "regular@example.com",
		}
		_, err := svc.CreatePost(req)
		require.NoError(t, err)
	})
}

// --- UpdatePost: non-admin can't set status to public ---

func TestUpdatePost_NonAdminCannotPublish(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIPostRepository(ctrl)
	svc := &postService{repository: repo, ctx: context.Background()}

	t.Run("non-admin cannot change status to public", func(t *testing.T) {
		repo.EXPECT().ExistsPost(1).Return(true)

		req := request.UpdatePostRequest{
			Id:     1,
			Status: enums.PostStatusPublic,
		}
		_, err := svc.UpdatePost(req, "regular@example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permiso")
	})

	t.Run("admin can change status to public", func(t *testing.T) {
		repo.EXPECT().ExistsPost(1).Return(true)
		repo.EXPECT().UpdatePost(uint(1), gomock.Any()).Return("Post actualizado", nil)

		req := request.UpdatePostRequest{
			Id:     1,
			Status: enums.PostStatusPublic,
		}
		_, err := svc.UpdatePost(req, "jsvaleriano321@gmail.com")
		require.NoError(t, err)
	})

	t.Run("non-admin can change status to private", func(t *testing.T) {
		repo.EXPECT().ExistsPost(1).Return(true)
		repo.EXPECT().UpdatePost(uint(1), gomock.Any()).Return("Post actualizado", nil)

		req := request.UpdatePostRequest{
			Id:     1,
			Status: enums.PostStatusPrivate,
		}
		_, err := svc.UpdatePost(req, "regular@example.com")
		require.NoError(t, err)
	})

	t.Run("invalid status rejected", func(t *testing.T) {
		repo.EXPECT().ExistsPost(1).Return(true)

		req := request.UpdatePostRequest{
			Id:     1,
			Status: "invalid",
		}
		_, err := svc.UpdatePost(req, "regular@example.com")
		assert.Error(t, err)
	})
}

// --- GetPendingPosts (service layer) ---

func TestGetPendingPostsService(t *testing.T) {
	db := setupTestDB(t)
	svc := NewPostService(context.Background(), nil, nil, db, nil, nil)

	// Create posts with different statuses
	db.Create(&models.Post{Title: "Pending Post", UserID: "u1", Status: enums.PostStatusPending})
	db.Create(&models.Post{Title: "Public Post", UserID: "u2", Status: enums.PostStatusPublic})
	db.Create(&models.Post{Title: "Private Post", UserID: "u1", Status: enums.PostStatusPrivate})
	db.Create(&models.Post{Title: "Another Pending", UserID: "u3", Status: enums.PostStatusPending})

	t.Run("returns only pending posts", func(t *testing.T) {
		pending, err := svc.GetPendingPosts()
		require.NoError(t, err)
		assert.Len(t, pending, 2)
		for _, p := range pending {
			assert.Equal(t, enums.PostStatusPending, p.Status)
		}
	})

	t.Run("no pending posts", func(t *testing.T) {
		cleanDB := setupTestDB(t)
		cleanSvc := NewPostService(context.Background(), nil, nil, cleanDB, nil, nil)
		pending, err := cleanSvc.GetPendingPosts()
		require.NoError(t, err)
		assert.Empty(t, pending)
	})
}

// --- ApprovePost ---

func TestApprovePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIPostRepository(ctrl)
	svc := &postService{repository: repo, ctx: context.Background()}

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().ExistsPost(1).Return(true)
		repo.EXPECT().UpdatePost(uint(1), gomock.Any()).DoAndReturn(func(id uint, p models.Post) (string, error) {
			assert.Equal(t, enums.PostStatusPublic, p.Status)
			return "Post actualizado", nil
		})

		result, err := svc.ApprovePost(1)
		require.NoError(t, err)
		assert.Contains(t, result, "actualizado")
	})

	t.Run("post not found", func(t *testing.T) {
		repo.EXPECT().ExistsPost(99).Return(false)

		_, err := svc.ApprovePost(99)
		assert.Error(t, err)
	})

	t.Run("update fails", func(t *testing.T) {
		repo.EXPECT().ExistsPost(1).Return(true)
		repo.EXPECT().UpdatePost(uint(1), gomock.Any()).Return("", errors.New("db error"))

		_, err := svc.ApprovePost(1)
		assert.Error(t, err)
	})
}

// --- RejectPost ---

func TestRejectPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockIPostRepository(ctrl)
	svc := &postService{repository: repo, ctx: context.Background()}

	t.Run("success", func(t *testing.T) {
		repo.EXPECT().ExistsPost(1).Return(true)
		repo.EXPECT().DeletePostById(1).Return(true)

		result, err := svc.RejectPost(1)
		require.NoError(t, err)
		assert.Contains(t, result, "rechazado")
	})

	t.Run("post not found", func(t *testing.T) {
		repo.EXPECT().ExistsPost(99).Return(false)

		_, err := svc.RejectPost(99)
		assert.Error(t, err)
	})

	t.Run("delete fails", func(t *testing.T) {
		repo.EXPECT().ExistsPost(1).Return(true)
		repo.EXPECT().DeletePostById(1).Return(false)

		_, err := svc.RejectPost(1)
		assert.Error(t, err)
	})
}

// --- NotificationService (mock-based) ---

func TestNotificationServiceWithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockINotificationRepository(ctrl)
	svc := NewNotificationService(mockRepo)

	t.Run("NotifyPendingPost saves notification", func(t *testing.T) {
		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.NotifyPendingPost(context.Background(), 1, "Test Post", "user1", "TestUser")
		require.NoError(t, err)
	})

	t.Run("NotifyPendingPost error", func(t *testing.T) {
		mockRepo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("redis error"))

		err := svc.NotifyPendingPost(context.Background(), 1, "Test Post", "user1", "TestUser")
		assert.Error(t, err)
	})

	t.Run("GetAll delegates to repo", func(t *testing.T) {
		mockRepo.EXPECT().GetAll(gomock.Any()).Return([]models.Notification{{ID: "1", Title: "Test"}}, nil)

		notifs, err := svc.GetAll(context.Background())
		require.NoError(t, err)
		assert.Len(t, notifs, 1)
	})

	t.Run("GetUnreadCount delegates to repo", func(t *testing.T) {
		mockRepo.EXPECT().GetUnreadCount(gomock.Any()).Return(int64(5), nil)

		count, err := svc.GetUnreadCount(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	t.Run("MarkRead delegates to repo", func(t *testing.T) {
		mockRepo.EXPECT().MarkRead(gomock.Any(), "id-1").Return(nil)

		err := svc.MarkRead(context.Background(), "id-1")
		require.NoError(t, err)
	})

	t.Run("MarkAllRead delegates to repo", func(t *testing.T) {
		mockRepo.EXPECT().MarkAllRead(gomock.Any()).Return(nil)

		err := svc.MarkAllRead(context.Background())
		require.NoError(t, err)
	})

	t.Run("Delete delegates to repo", func(t *testing.T) {
		mockRepo.EXPECT().Delete(gomock.Any(), "id-1").Return(nil)

		err := svc.Delete(context.Background(), "id-1")
		require.NoError(t, err)
	})
}