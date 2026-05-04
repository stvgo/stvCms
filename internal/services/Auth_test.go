package services

import (
	"testing"

	"stvCms/internal/models"
	"stvCms/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}))
	return db
}

func TestAuthService_SyncUser(t *testing.T) {
	t.Run("crea nuevo usuario", func(t *testing.T) {
		repo := repository.NewUserRepository(setupAuthTestDB(t))
		svc := NewAuthService(repo)

		user, err := svc.SyncUser("test@test.com", "Test User", "http://img.com/pic.jpg", "google123")
		require.NoError(t, err)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "test@test.com", user.Email)
		assert.Equal(t, "google123", user.GoogleID)
		assert.Equal(t, "user", user.Role)
	})

	t.Run("actualiza usuario existente", func(t *testing.T) {
		db := setupAuthTestDB(t)
		repo := repository.NewUserRepository(db)
		svc := NewAuthService(repo)

		_, err := svc.SyncUser("test@test.com", "Original", "http://img.com/old.jpg", "g1")
		require.NoError(t, err)

		user, err := svc.SyncUser("test@test.com", "Updated", "http://img.com/new.jpg", "g2")
		require.NoError(t, err)
		assert.Equal(t, "Updated", user.Name)
		assert.Equal(t, "http://img.com/new.jpg", user.Image)
		assert.Equal(t, "g2", user.GoogleID)
	})

	t.Run("no sobreescribe googleID vacio", func(t *testing.T) {
		db := setupAuthTestDB(t)
		repo := repository.NewUserRepository(db)
		svc := NewAuthService(repo)

		_, err := svc.SyncUser("test@test.com", "Test", "", "g1")
		require.NoError(t, err)

		user, err := svc.SyncUser("test@test.com", "Updated", "", "")
		require.NoError(t, err)
		assert.Equal(t, "g1", user.GoogleID)
	})
}

func TestAuthService_GetUserByEmail(t *testing.T) {
	t.Run("encontrado", func(t *testing.T) {
		db := setupAuthTestDB(t)
		repo := repository.NewUserRepository(db)
		svc := NewAuthService(repo)

		_, err := svc.SyncUser("find@test.com", "Find Me", "", "")
		require.NoError(t, err)

		user, err := svc.GetUserByEmail("find@test.com")
		require.NoError(t, err)
		assert.Equal(t, "Find Me", user.Name)
	})

	t.Run("no encontrado", func(t *testing.T) {
		repo := repository.NewUserRepository(setupAuthTestDB(t))
		svc := NewAuthService(repo)
		_, err := svc.GetUserByEmail("ghost@test.com")
		assert.Error(t, err)
	})
}