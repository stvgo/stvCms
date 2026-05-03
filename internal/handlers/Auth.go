package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"stvCms/internal/middleware"
	"stvCms/internal/services"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	service services.IAuthService
}

func NewAuthHandler(service services.IAuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type GoogleLoginRequest struct {
	Credential string `json:"credential"`
}

type GoogleLoginResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

type UserProfile struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Image string `json:"image"`
	Role  string `json:"role"`
}

type googleTokenInfo struct {
	Iss           string `json:"iss"`
	Sub           string `json:"sub"`
	Aud           string `json:"aud"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	var req GoogleLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if req.Credential == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "credential is required"})
	}

	info, err := verifyGoogleIDToken(req.Credential)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	user, err := h.service.SyncUser(info.Email, info.Name, info.Picture, info.Sub)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to sync user"})
	}

	token, err := middleware.GenerateToken(fmt.Sprintf("%d", user.ID), user.Email, user.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to generate token"})
	}

	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, GoogleLoginResponse{
		Token: token,
		User: UserProfile{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
			Image: user.Image,
			Role:  user.Role,
		},
	})
}

func verifyGoogleIDToken(idToken string) (*googleTokenInfo, error) {
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=%s", idToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid token")
	}

	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode token info: %w", err)
	}

	clientID := os.Getenv("CLIENT_ID")
	if info.Aud != clientID {
		return nil, fmt.Errorf("invalid audience")
	}

	if info.Iss != "https://accounts.google.com" && info.Iss != "accounts.google.com" {
		return nil, fmt.Errorf("invalid issuer")
	}

	if info.Email == "" || info.EmailVerified != "true" {
		return nil, fmt.Errorf("email not verified")
	}

	return &info, nil
}

func (h *AuthHandler) Me(c echo.Context) error {
	email, ok := c.Get("email").(string)
	if !ok || email == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	user, err := h.service.GetUserByEmail(email)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"image": user.Image,
		"role":  user.Role,
	})
}
