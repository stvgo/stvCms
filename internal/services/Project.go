package services

import (
	"fmt"
	"stvCms/internal/models"
	"stvCms/internal/repository"
	"stvCms/internal/rest/request"
	"stvCms/internal/rest/response"
	"stvCms/internal/services/enums"

	"gorm.io/gorm"
)

type IProjectService interface {
	CreateProject(req request.CreateProjectRequest) (string, error)
	GetProjects(userID string) ([]response.ProjectResponse, error)
	GetPublicProjects() ([]response.ProjectResponse, error)
	GetProjectByID(id int) (response.ProjectResponse, error)
	UpdateProject(req request.UpdateProjectRequest) (string, error)
	DeleteProjectByID(id string) (string, error)
}

type projectService struct {
	repository repository.IProjectRepository
}

func NewProjectService(db *gorm.DB) *projectService {
	return &projectService{
		repository: repository.NewProjectGormRepository(db),
	}
}

func (ps *projectService) CreateProject(req request.CreateProjectRequest) (string, error) {
	projectType := req.Type
	validTypes := map[string]bool{
		enums.ProjectTypeGame:    true,
		enums.ProjectTypeWeb:     true,
		enums.ProjectTypeAPI:     true,
		enums.ProjectTypeTool:    true,
		enums.ProjectTypeLibrary: true,
	}
	if !validTypes[projectType] {
		return "", fmt.Errorf("tipo de project inválido: %s", projectType)
	}

	project := models.Project{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		URL:         req.URL,
		EmbedURL:    req.EmbedURL,
		ImageURL:    req.ImageURL,
		GitHubURL:   req.GitHubURL,
		TechStack:   req.TechStack,
		UserID:      req.UserID,
	}

	result, err := ps.repository.CreateProject(project)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (ps *projectService) GetProjects(userID string) ([]response.ProjectResponse, error) {
	projects, err := ps.repository.GetProjects(userID)
	if err != nil {
		return nil, err
	}

	var responses []response.ProjectResponse
	for _, p := range projects {
		responses = append(responses, toProjectResponse(p))
	}
	return responses, nil
}

func (ps *projectService) GetPublicProjects() ([]response.ProjectResponse, error) {
	projects, err := ps.repository.GetPublicProjects()
	if err != nil {
		return nil, err
	}

	var responses []response.ProjectResponse
	for _, p := range projects {
		responses = append(responses, toProjectResponse(p))
	}
	return responses, nil
}

func (ps *projectService) GetProjectByID(id int) (response.ProjectResponse, error) {
	project, err := ps.repository.GetProjectById(uint(id))
	if err != nil {
		return response.ProjectResponse{}, err
	}
	return toProjectResponse(project), nil
}

func (ps *projectService) UpdateProject(req request.UpdateProjectRequest) (string, error) {
	if !ps.repository.ExistsProject(int(req.Id)) {
		return "", gorm.ErrRecordNotFound
	}

	if req.Type != "" {
		validTypes := map[string]bool{
			enums.ProjectTypeGame:    true,
			enums.ProjectTypeWeb:     true,
			enums.ProjectTypeAPI:     true,
			enums.ProjectTypeTool:    true,
			enums.ProjectTypeLibrary: true,
		}
		if !validTypes[req.Type] {
			return "", fmt.Errorf("tipo de project inválido: %s", req.Type)
		}
	}

	project := models.Project{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		URL:         req.URL,
		EmbedURL:    req.EmbedURL,
		ImageURL:    req.ImageURL,
		GitHubURL:   req.GitHubURL,
		TechStack:   req.TechStack,
	}

	result, err := ps.repository.UpdateProject(req.Id, project)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (ps *projectService) DeleteProjectByID(id string) (string, error) {
	projectId := 0
	fmt.Sscanf(id, "%d", &projectId)

	if !ps.repository.DeleteProjectById(projectId) {
		return "", fmt.Errorf("error al borrar el project")
	}
	return "Project borrado", nil
}

func toProjectResponse(p models.Project) response.ProjectResponse {
	return response.ProjectResponse{
		Id:          p.ID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		Title:       p.Title,
		Description: p.Description,
		Type:        p.Type,
		URL:         p.URL,
		EmbedURL:    p.EmbedURL,
		ImageURL:    p.ImageURL,
		GitHubURL:   p.GitHubURL,
		TechStack:   p.TechStack,
		UserID:      p.UserID,
	}
}