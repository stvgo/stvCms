package repository

import (
	"testing"

	"stvCms/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}))
	return db
}

func TestUserRepository_FindByEmail(t *testing.T) {
	t.Run("encontrado", func(t *testing.T) {
		db := setupUserTestDB(t)
		repo := NewUserRepository(db)
		user := &models.User{Email: "test@test.com", Name: "Test", GoogleID: "g1", Role: "user"}
		require.NoError(t, repo.Create(user))

		found, err := repo.FindByEmail("test@test.com")
		require.NoError(t, err)
		assert.Equal(t, "Test", found.Name)
		assert.Equal(t, "test@test.com", found.Email)
	})

	t.Run("no encontrado", func(t *testing.T) {
		repo := NewUserRepository(setupUserTestDB(t))
		_, err := repo.FindByEmail("nonexistent@test.com")
		assert.Error(t, err)
	})
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("crea usuario", func(t *testing.T) {
		repo := NewUserRepository(setupUserTestDB(t))
		user := &models.User{Email: "new@test.com", Name: "New", GoogleID: "g2", Role: "user"}
		err := repo.Create(user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Run("actualiza usuario", func(t *testing.T) {
		db := setupUserTestDB(t)
		repo := NewUserRepository(db)
		user := &models.User{Email: "up@test.com", Name: "Up", GoogleID: "g3", Role: "user"}
		require.NoError(t, repo.Create(user))

		user.Name = "Updated"
		require.NoError(t, repo.Update(user))

		found, err := repo.FindByEmail("up@test.com")
		require.NoError(t, err)
		assert.Equal(t, "Updated", found.Name)
	})
}