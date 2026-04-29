package services

import (
	"stvCms/internal/models"
	"stvCms/internal/repository"
)

type IAuthService interface {
	SyncUser(email, name, image, googleId string) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
}

type authService struct {
	repo repository.IUserRepository
}

func NewAuthService(repo repository.IUserRepository) IAuthService {
	return &authService{repo: repo}
}

func (s *authService) SyncUser(email, name, image, googleId string) (models.User, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		// Crear nuevo usuario
		newUser := models.User{
			Email:    email,
			Name:     name,
			Image:    image,
			GoogleID: googleId,
			Role:     "user",
		}
		if err := s.repo.Create(&newUser); err != nil {
			return models.User{}, err
		}
		return newUser, nil
	}

	// Actualizar usuario existente
	user.Name = name
	user.Image = image
	if googleId != "" {
		user.GoogleID = googleId
	}
	if err := s.repo.Update(&user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *authService) GetUserByEmail(email string) (models.User, error) {
	return s.repo.FindByEmail(email)
}
