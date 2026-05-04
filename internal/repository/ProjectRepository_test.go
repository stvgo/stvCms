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

func setupProjectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Project{}))
	return db
}

func seedProject(t *testing.T, db *gorm.DB, title, projectType, userID string) models.Project {
	t.Helper()
	project := models.Project{Title: title, Type: projectType, UserID: userID}
	require.NoError(t, db.Create(&project).Error)
	return project
}

// --- CreateProject ---

func TestRepo_CreateProject(t *testing.T) {
	t.Run("project simple", func(t *testing.T) {
		repo := NewProjectGormRepository(setupProjectTestDB(t))
		_, err := repo.CreateProject(models.Project{Title: "My Game", Type: "game", UserID: "u1"})
		require.NoError(t, err)
	})

	t.Run("project con todos los campos", func(t *testing.T) {
		db := setupProjectTestDB(t)
		repo := NewProjectGormRepository(db)
		project := models.Project{
			Title:       "Flappy Bird",
			Description: "A fun game",
			Type:        "game",
			URL:         "https://example.com",
			EmbedURL:    "https://example.com/embed",
			ImageURL:    "https://example.com/img.png",
			GitHubURL:   "https://github.com/user/repo",
			TechStack:   "Python,Pygame",
			UserID:      "u1",
		}
		msg, err := repo.CreateProject(project)
		require.NoError(t, err)
		assert.Equal(t, "Project creado", msg)

		var count int64
		db.Model(&models.Project{}).Count(&count)
		assert.Equal(t, int64(1), count)
	})
}

// --- GetProjects ---

func TestRepo_GetProjects(t *testing.T) {
	t.Run("lista vacía", func(t *testing.T) {
		repo := NewProjectGormRepository(setupProjectTestDB(t))
		projects, err := repo.GetProjects("u1")
		require.NoError(t, err)
		assert.Empty(t, projects)
	})

	t.Run("solo projects del usuario", func(t *testing.T) {
		db := setupProjectTestDB(t)
		repo := NewProjectGormRepository(db)
		seedProject(t, db, "Game 1", "game", "u1")
		seedProject(t, db, "Game 2", "game", "u1")
		seedProject(t, db, "Web 1", "web", "u2")

		projects, err := repo.GetProjects("u1")
		require.NoError(t, err)
		assert.Len(t, projects, 2)

		projects2, err := repo.GetProjects("u2")
		require.NoError(t, err)
		assert.Len(t, projects2, 1)
		assert.Equal(t, "Web 1", projects2[0].Title)
	})
}

// --- GetPublicProjects ---

func TestRepo_GetPublicProjects(t *testing.T) {
	t.Run("todos los projects son públicos", func(t *testing.T) {
		db := setupProjectTestDB(t)
		repo := NewProjectGormRepository(db)
		seedProject(t, db, "Game 1", "game", "u1")
		seedProject(t, db, "Web 1", "web", "u2")

		projects, err := repo.GetPublicProjects()
		require.NoError(t, err)
		assert.Len(t, projects, 2)
	})

	t.Run("lista vacía", func(t *testing.T) {
		repo := NewProjectGormRepository(setupProjectTestDB(t))
		projects, err := repo.GetPublicProjects()
		require.NoError(t, err)
		assert.Empty(t, projects)
	})
}

// --- GetProjectById ---

func TestRepo_GetProjectById(t *testing.T) {
	t.Run("encontrado", func(t *testing.T) {
		db := setupProjectTestDB(t)
		repo := NewProjectGormRepository(db)
		saved := seedProject(t, db, "My Game", "game", "u1")

		project, err := repo.GetProjectById(saved.ID)
		require.NoError(t, err)
		assert.Equal(t, "My Game", project.Title)
		assert.Equal(t, "game", project.Type)
		assert.Equal(t, "u1", project.UserID)
	})

	t.Run("no encontrado", func(t *testing.T) {
		repo := NewProjectGormRepository(setupProjectTestDB(t))
		_, err := repo.GetProjectById(9999)
		assert.Error(t, err)
	})
}

// --- UpdateProject ---

func TestRepo_UpdateProject(t *testing.T) {
	t.Run("actualiza campos", func(t *testing.T) {
		db := setupProjectTestDB(t)
		repo := NewProjectGormRepository(db)
		saved := seedProject(t, db, "Original", "game", "u1")

		msg, err := repo.UpdateProject(saved.ID, models.Project{Title: "Updated", Type: "web"})
		require.NoError(t, err)
		assert.Equal(t, "Project actualizado", msg)

		updated, _ := repo.GetProjectById(saved.ID)
		assert.Equal(t, "Updated", updated.Title)
		assert.Equal(t, "web", updated.Type)
	})

	t.Run("ID no existe", func(t *testing.T) {
		repo := NewProjectGormRepository(setupProjectTestDB(t))
		msg, err := repo.UpdateProject(9999, models.Project{Title: "Ghost"})
		require.NoError(t, err)
		assert.Contains(t, msg, "no habían datos")
	})
}

// --- DeleteProjectById ---

func TestRepo_DeleteProjectById(t *testing.T) {
	t.Run("elimina", func(t *testing.T) {
		db := setupProjectTestDB(t)
		repo := NewProjectGormRepository(db)
		saved := seedProject(t, db, "To delete", "game", "u1")

		ok := repo.DeleteProjectById(int(saved.ID))
		assert.True(t, ok)
		assert.False(t, repo.ExistsProject(int(saved.ID)))
	})

	t.Run("no existe", func(t *testing.T) {
		repo := NewProjectGormRepository(setupProjectTestDB(t))
		ok := repo.DeleteProjectById(9999)
		assert.False(t, ok)
	})
}

// --- ExistsProject ---

func TestRepo_ExistsProject(t *testing.T) {
	t.Run("existe", func(t *testing.T) {
		db := setupProjectTestDB(t)
		repo := NewProjectGormRepository(db)
		saved := seedProject(t, db, "Present", "game", "u1")
		assert.True(t, repo.ExistsProject(int(saved.ID)))
	})

	t.Run("no existe", func(t *testing.T) {
		repo := NewProjectGormRepository(setupProjectTestDB(t))
		assert.False(t, repo.ExistsProject(9999))
	})
}

// --- Error paths (DB connection closed) ---

func TestRepo_ProjectErrorPaths(t *testing.T) {
	closedProjectDB := func(t *testing.T) *gorm.DB {
		t.Helper()
		db := setupProjectTestDB(t)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.Close()
		return db
	}

	t.Run("CreateProject con DB cerrada", func(t *testing.T) {
		repo := NewProjectGormRepository(closedProjectDB(t))
		msg, err := repo.CreateProject(models.Project{Title: "fail", Type: "game"})
		assert.Error(t, err)
		assert.Equal(t, "No se pudo crear el project", msg)
	})

	t.Run("GetProjects con DB cerrada", func(t *testing.T) {
		repo := NewProjectGormRepository(closedProjectDB(t))
		_, err := repo.GetProjects("u1")
		assert.Error(t, err)
	})

	t.Run("GetPublicProjects con DB cerrada", func(t *testing.T) {
		repo := NewProjectGormRepository(closedProjectDB(t))
		_, err := repo.GetPublicProjects()
		assert.Error(t, err)
	})

	t.Run("GetProjectById con DB cerrada", func(t *testing.T) {
		repo := NewProjectGormRepository(closedProjectDB(t))
		_, err := repo.GetProjectById(1)
		assert.Error(t, err)
	})

	t.Run("UpdateProject con DB cerrada", func(t *testing.T) {
		repo := NewProjectGormRepository(closedProjectDB(t))
		msg, err := repo.UpdateProject(1, models.Project{Title: "fail"})
		assert.Error(t, err)
		assert.Equal(t, "No se pudo actualizar el project", msg)
	})
}