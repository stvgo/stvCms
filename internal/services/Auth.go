package services

import (
	"stvCms/internal/models"
	"stvCms/internal/repository"
)

type IAuthService interface {
	SyncUser(email, name, image, googleID string) (models.User, error)
	SyncUserWithGitHub(email, name, image, githubID string) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
}

type authService struct {
	repo repository.IUserRepository
}

func NewAuthService(repo repository.IUserRepository) IAuthService {
	return &authService{repo: repo}
}

func (s *authService) SyncUser(email, name, image, googleID string) (models.User, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		newUser := models.User{
			Email:    email,
			Name:     name,
			Image:    image,
			GoogleID: googleID,
			Role:     "user",
		}
		if err := s.repo.Create(&newUser); err != nil {
			return models.User{}, err
		}
		return newUser, nil
	}

	user.Name = name
	user.Image = image
	if googleID != "" {
		user.GoogleID = googleID
	}
	if err := s.repo.Update(&user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *authService) SyncUserWithGitHub(email, name, image, githubID string) (models.User, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		newUser := models.User{
			Email:    email,
			Name:     name,
			Image:    image,
			GitHubID: githubID,
			Role:     "user",
		}
		if err := s.repo.Create(&newUser); err != nil {
			return models.User{}, err
		}
		return newUser, nil
	}

	user.Name = name
	user.Image = image
	if githubID != "" {
		user.GitHubID = githubID
	}
	if err := s.repo.Update(&user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *authService) GetUserByEmail(email string) (models.User, error) {
	return s.repo.FindByEmail(email)
}
