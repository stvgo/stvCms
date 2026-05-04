package repository

import (
	"stvCms/internal/models"

	"gorm.io/gorm"
)

type IProjectRepository interface {
	CreateProject(project models.Project) (string, error)
	GetProjects(userID string) ([]models.Project, error)
	GetPublicProjects() ([]models.Project, error)
	GetProjectById(id uint) (models.Project, error)
	UpdateProject(id uint, project models.Project) (string, error)
	DeleteProjectById(id int) bool
	ExistsProject(id int) bool
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectGormRepository(db *gorm.DB) *projectRepository {
	return &projectRepository{db: db}
}

func (pr *projectRepository) CreateProject(project models.Project) (string, error) {
	err := pr.db.Create(&project).Error
	if err != nil {
		return "No se pudo crear el project", err
	}
	return "Project creado", nil
}

func (pr *projectRepository) GetProjects(userID string) ([]models.Project, error) {
	var projects []models.Project
	err := pr.db.Where("user_id = ?", userID).Find(&projects).Error
	if err != nil {
		return projects, err
	}
	return projects, nil
}

func (pr *projectRepository) GetPublicProjects() ([]models.Project, error) {
	var projects []models.Project
	err := pr.db.Find(&projects).Error
	if err != nil {
		return projects, err
	}
	return projects, nil
}

func (pr *projectRepository) GetProjectById(id uint) (models.Project, error) {
	var project models.Project
	err := pr.db.First(&project, id).Error
	return project, err
}

func (pr *projectRepository) UpdateProject(id uint, project models.Project) (string, error) {
	result := pr.db.Model(&models.Project{}).Where("id = ?", id).Updates(project)
	if result.Error != nil {
		return "No se pudo actualizar el project", result.Error
	}
	if result.RowsAffected == 0 {
		return "El project no fue encontrado o no habían datos para actualizar", nil
	}
	return "Project actualizado", nil
}

func (pr *projectRepository) DeleteProjectById(id int) bool {
	return pr.db.Delete(&models.Project{}, id).RowsAffected > 0
}

func (pr *projectRepository) ExistsProject(id int) bool {
	err := pr.db.First(&models.Project{}, id).Error
	return err == nil
}