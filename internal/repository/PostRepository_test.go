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

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Post{}, &models.ContentBlock{}))
	return db
}

func seedPost(t *testing.T, db *gorm.DB, title, userID string, blocks ...models.ContentBlock) models.Post {
	t.Helper()
	post := models.Post{Title: title, UserID: userID, IsVisible: true, ContentBlocks: blocks}
	require.NoError(t, db.Create(&post).Error)
	return post
}

// --- CreatePost ---

func TestRepo_CreatePost(t *testing.T) {
	t.Run("post simple", func(t *testing.T) {
		repo := NewPostGormRepository(setupTestDB(t))
		_, err := repo.CreatePost(models.Post{Title: "Hello", UserID: "u1"})
		require.NoError(t, err)
	})

	t.Run("post con content blocks", func(t *testing.T) {
		repo := NewPostGormRepository(setupTestDB(t))
		post := models.Post{
			Title:  "With blocks",
			UserID: "u1",
			ContentBlocks: []models.ContentBlock{
				{Type: "text", Order: 1, Content: "body"},
				{Type: "code", Order: 2, Content: "main()", Language: "go"},
			},
		}
		msg, err := repo.CreatePost(post)
		require.NoError(t, err)
		assert.Equal(t, "Post creado", msg)
	})
}

// --- GetPosts ---

func TestRepo_GetPosts(t *testing.T) {
	t.Run("lista vacía", func(t *testing.T) {
		repo := NewPostGormRepository(setupTestDB(t))
		posts, err := repo.GetPosts()
		require.NoError(t, err)
		assert.Empty(t, posts)
	})

	t.Run("múltiples posts con preload", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostGormRepository(db)
		seedPost(t, db, "Post 1", "u1", models.ContentBlock{Type: "text", Order: 1, Content: "a"})
		seedPost(t, db, "Post 2", "u2")

		posts, err := repo.GetPosts()
		require.NoError(t, err)
		assert.Len(t, posts, 2)
		assert.Len(t, posts[0].ContentBlocks, 1)
	})
}

// --- GetPostById ---

func TestRepo_GetPostById(t *testing.T) {
	t.Run("encontrado con content blocks", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostGormRepository(db)
		saved := seedPost(t, db, "Detail", "u1",
			models.ContentBlock{Type: "code", Order: 1, Content: "fmt.Println()", Language: "go"},
		)

		post, err := repo.GetPostById(saved.ID)
		require.NoError(t, err)
		assert.Equal(t, "Detail", post.Title)
		assert.Len(t, post.ContentBlocks, 1)
		assert.Equal(t, "go", post.ContentBlocks[0].Language)
	})

	t.Run("no encontrado", func(t *testing.T) {
		repo := NewPostGormRepository(setupTestDB(t))
		_, err := repo.GetPostById(9999)
		assert.Error(t, err)
	})
}

// --- GetPostsByFilter ---

func TestRepo_GetPostsByFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostGormRepository(db)
	seedPost(t, db, "Go tutorial", "user-go")
	seedPost(t, db, "Python guide", "user-py",
		models.ContentBlock{Type: "text", Order: 1, Content: "learn python fast"},
	)

	t.Run("por título", func(t *testing.T) {
		posts, err := repo.GetPostsByFilter("Go")
		require.NoError(t, err)
		assert.Len(t, posts, 1)
		assert.Equal(t, "Go tutorial", posts[0].Title)
	})

	t.Run("por userID", func(t *testing.T) {
		posts, err := repo.GetPostsByFilter("user-py")
		require.NoError(t, err)
		assert.Len(t, posts, 1)
	})

	t.Run("por contenido de content block", func(t *testing.T) {
		posts, err := repo.GetPostsByFilter("learn python")
		require.NoError(t, err)
		assert.Len(t, posts, 1)
	})

	t.Run("sin resultados", func(t *testing.T) {
		posts, err := repo.GetPostsByFilter("xyz-nonexistent")
		require.NoError(t, err)
		assert.Empty(t, posts)
	})
}

// --- UpdatePost ---

func TestRepo_UpdatePost(t *testing.T) {
	t.Run("actualiza campos", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostGormRepository(db)
		saved := seedPost(t, db, "Original", "u1")

		msg, err := repo.UpdatePost(saved.ID, models.Post{Title: "Updated"})
		require.NoError(t, err)
		assert.Equal(t, "Post actualizado", msg)

		updated, _ := repo.GetPostById(saved.ID)
		assert.Equal(t, "Updated", updated.Title)
	})

	t.Run("ID no existe — sin rows afectadas", func(t *testing.T) {
		repo := NewPostGormRepository(setupTestDB(t))
		msg, err := repo.UpdatePost(9999, models.Post{Title: "Ghost"})
		require.NoError(t, err)
		assert.Contains(t, msg, "no habían datos")
	})
}

// --- DeletePostById ---

func TestRepo_DeletePostById(t *testing.T) {
	t.Run("elimina", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostGormRepository(db)
		saved := seedPost(t, db, "To delete", "u1")

		ok := repo.DeletePostById(int(saved.ID))
		assert.True(t, ok)

		assert.False(t, repo.ExistsPost(int(saved.ID)))
	})

	t.Run("no existe", func(t *testing.T) {
		repo := NewPostGormRepository(setupTestDB(t))
		ok := repo.DeletePostById(9999)
		assert.False(t, ok)
	})
}

// --- ExistsPost ---

func TestRepo_ExistsPost(t *testing.T) {
	t.Run("existe", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostGormRepository(db)
		saved := seedPost(t, db, "Present", "u1")
		assert.True(t, repo.ExistsPost(int(saved.ID)))
	})

	t.Run("no existe", func(t *testing.T) {
		repo := NewPostGormRepository(setupTestDB(t))
		assert.False(t, repo.ExistsPost(9999))
	})
}

// --- Error paths (DB connection closed) ---

func closedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()
	return db
}

func TestRepo_ErrorPaths(t *testing.T) {
	t.Run("CreatePost con DB cerrada", func(t *testing.T) {
		repo := NewPostGormRepository(closedDB(t))
		msg, err := repo.CreatePost(models.Post{Title: "fail"})
		assert.Error(t, err)
		assert.Equal(t, "No se pudo crear el post", msg)
	})

	t.Run("GetPosts con DB cerrada", func(t *testing.T) {
		repo := NewPostGormRepository(closedDB(t))
		_, err := repo.GetPosts()
		assert.Error(t, err)
	})

	t.Run("GetPostsByFilter con DB cerrada", func(t *testing.T) {
		repo := NewPostGormRepository(closedDB(t))
		_, err := repo.GetPostsByFilter("anything")
		assert.Error(t, err)
	})

	t.Run("UpdatePost con DB cerrada", func(t *testing.T) {
		repo := NewPostGormRepository(closedDB(t))
		msg, err := repo.UpdatePost(1, models.Post{Title: "fail"})
		assert.Error(t, err)
		assert.Equal(t, "No se pudo actualizar el post", msg)
	})
}
