package handlers

import (
	"net/http"
	"strconv"

	"stvCms/internal/rest/request"
	"stvCms/internal/services"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type projectHandler struct {
	service services.IProjectService
}

func NewProjectHandler(db *gorm.DB) *projectHandler {
	return &projectHandler{
		service: services.NewProjectService(db),
	}
}

func (h *projectHandler) CreateProject(c echo.Context) error {
	var input request.CreateProjectRequest
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	input.UserID = getUserName(c)

	result, err := h.service.CreateProject(input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": result})
}

func (h *projectHandler) GetProjects(c echo.Context) error {
	userID := getUserName(c)

	projects, err := h.service.GetProjects(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, projects)
}

func (h *projectHandler) GetPublicProjects(c echo.Context) error {
	projects, err := h.service.GetPublicProjects()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, projects)
}

func (h *projectHandler) GetProjectById(c echo.Context) error {
	id := c.Param("id")

	projectId, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}

	project, err := h.service.GetProjectByID(projectId)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Project not found"})
	}

	return c.JSON(http.StatusOK, project)
}

func (h *projectHandler) UpdateProject(c echo.Context) error {
	var input request.UpdateProjectRequest
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	result, err := h.service.UpdateProject(input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": result})
}

func (h *projectHandler) DeleteProjectById(c echo.Context) error {
	id := c.Param("id")

	_, err := h.service.DeleteProjectByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusNoContent, echo.Map{"message": "Deleted"})
}