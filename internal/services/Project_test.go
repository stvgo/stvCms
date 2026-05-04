package services

import (
	"testing"

	"stvCms/internal/models"
	"stvCms/internal/rest/request"
	"stvCms/internal/services/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupProjectServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Project{}))
	return db
}

// --- CreateProject ---

func TestCreateProject(t *testing.T) {
	t.Run("éxito con campos completos", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		req := request.CreateProjectRequest{
			Title:       "Flappy Bird",
			Description: "A fun game",
			Type:        enums.ProjectTypeGame,
			URL:         "https://example.com",
			EmbedURL:    "https://example.com/embed",
			GitHubURL:   "https://github.com/user/repo",
			TechStack:   "Python,Pygame",
			UserID:      "u1",
		}
		msg, err := svc.CreateProject(req)
		require.NoError(t, err)
		assert.Equal(t, "Project creado", msg)
	})

	t.Run("éxito con campos mínimos", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		req := request.CreateProjectRequest{
			Title:  "Minimal",
			Type:   enums.ProjectTypeWeb,
			UserID: "u1",
		}
		msg, err := svc.CreateProject(req)
		require.NoError(t, err)
		assert.Equal(t, "Project creado", msg)
	})

	t.Run("tipo inválido", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		req := request.CreateProjectRequest{
			Title:  "Bad Type",
			Type:   "invalid",
			UserID: "u1",
		}
		_, err := svc.CreateProject(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inválido")
	})

	t.Run("todos los tipos válidos", func(t *testing.T) {
		types := []string{enums.ProjectTypeGame, enums.ProjectTypeWeb, enums.ProjectTypeAPI, enums.ProjectTypeTool, enums.ProjectTypeLibrary}
		for _, typ := range types {
			svc := NewProjectService(setupProjectServiceTestDB(t))
			req := request.CreateProjectRequest{
				Title:  "Project " + typ,
				Type:   typ,
				UserID: "u1",
			}
			msg, err := svc.CreateProject(req)
			require.NoError(t, err, "type %s should be valid", typ)
			assert.Equal(t, "Project creado", msg)
		}
	})
}

// --- GetProjects ---

func TestGetProjects(t *testing.T) {
	t.Run("retorna projects del usuario", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Game 1", Type: enums.ProjectTypeGame, UserID: "u1"})
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Game 2", Type: enums.ProjectTypeGame, UserID: "u1"})
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Web 1", Type: enums.ProjectTypeWeb, UserID: "u2"})

		projects, err := svc.GetProjects("u1")
		require.NoError(t, err)
		assert.Len(t, projects, 2)
	})

	t.Run("usuario sin projects", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		projects, err := svc.GetProjects("u1")
		require.NoError(t, err)
		assert.Empty(t, projects)
	})
}

// --- GetPublicProjects ---

func TestGetPublicProjects(t *testing.T) {
	t.Run("retorna todos los projects", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Game 1", Type: enums.ProjectTypeGame, UserID: "u1"})
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Web 1", Type: enums.ProjectTypeWeb, UserID: "u2"})

		projects, err := svc.GetPublicProjects()
		require.NoError(t, err)
		assert.Len(t, projects, 2)
	})
}

// --- GetProjectByID ---

func TestGetProjectByID(t *testing.T) {
	t.Run("encontrado", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "My Game", Type: enums.ProjectTypeGame, UserID: "u1", GitHubURL: "https://github.com/test", TechStack: "Go"})

		project, err := svc.GetProjectByID(1)
		require.NoError(t, err)
		assert.Equal(t, "My Game", project.Title)
		assert.Equal(t, "game", project.Type)
		assert.Equal(t, "https://github.com/test", project.GitHubURL)
	})

	t.Run("no encontrado", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, err := svc.GetProjectByID(9999)
		assert.Error(t, err)
	})
}

// --- UpdateProject ---

func TestUpdateProject(t *testing.T) {
	t.Run("éxito actualizando campos", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Original", Type: enums.ProjectTypeGame, UserID: "u1"})

		msg, err := svc.UpdateProject(request.UpdateProjectRequest{
			Id:    1,
			Title: stringPtr("Updated"),
			Type:  stringPtr(enums.ProjectTypeWeb),
		})
		require.NoError(t, err)
		assert.Equal(t, "Project actualizado", msg)

		project, _ := svc.GetProjectByID(1)
		assert.Equal(t, "Updated", project.Title)
		assert.Equal(t, "web", project.Type)
	})

	t.Run("project no existe", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, err := svc.UpdateProject(request.UpdateProjectRequest{Id: 9999, Title: stringPtr("Ghost")})
		assert.Error(t, err)
	})

	t.Run("tipo inválido", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Original", Type: enums.ProjectTypeGame, UserID: "u1"})

		_, err := svc.UpdateProject(request.UpdateProjectRequest{Id: 1, Type: stringPtr("invalid")})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "inválido")
	})
}

// --- DeleteProjectByID ---

func TestDeleteProjectByID(t *testing.T) {
	t.Run("éxito", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "To delete", Type: enums.ProjectTypeGame, UserID: "u1"})

		msg, err := svc.DeleteProjectByID("1")
		require.NoError(t, err)
		assert.Equal(t, "Project borrado", msg)

		_, err = svc.GetProjectByID(1)
		assert.Error(t, err)
	})

	t.Run("ID inválido", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, err := svc.DeleteProjectByID("abc")
		assert.Error(t, err)
	})

	t.Run("no encontrado", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, err := svc.DeleteProjectByID("9999")
		assert.Error(t, err)
	})
}

// helper para crear punteros a strings, ya que UpdateProjectRequest usa campos opcionales
// pero GORM Updates ignora zero values, así que pasamos strings directamente cuando queremos actualizar
func stringPtr(s string) string {
	return s
}

// --- Coverage: GetProjectByID via service for public endpoint ---

func TestGetPublicProjectByID(t *testing.T) {
	t.Run("encontrado vía GetPublicProjects", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Public Game", Type: enums.ProjectTypeWeb, UserID: "u1"})

		projects, err := svc.GetPublicProjects()
		require.NoError(t, err)
		require.Len(t, projects, 1)
		assert.Equal(t, "Public Game", projects[0].Title)
		assert.Equal(t, "u1", projects[0].UserID)
	})
}

// --- DeleteProjectByID edge cases ---

func TestDeleteProjectByID_EdgeCases(t *testing.T) {
	t.Run("borrar project y verificar desapareció", func(t *testing.T) {
		svc := NewProjectService(setupProjectServiceTestDB(t))
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Game 1", Type: enums.ProjectTypeGame, UserID: "u1"})
		_, _ = svc.CreateProject(request.CreateProjectRequest{Title: "Game 2", Type: enums.ProjectTypeGame, UserID: "u1"})

		msg, err := svc.DeleteProjectByID("1")
		require.NoError(t, err)
		assert.Equal(t, "Project borrado", msg)

		projects, _ := svc.GetProjects("u1")
		assert.Len(t, projects, 1)
		assert.Equal(t, "Game 2", projects[0].Title)
	})
}