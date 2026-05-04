package handlers

import (
	"github.com/labstack/echo/v4"
)

type projectsHandler struct{}

func NewProjectsHandler() *projectsHandler {
	return &projectsHandler{}
}

func (p *projectsHandler) ViewProjectsHandler(e echo.Context) {
	
}
