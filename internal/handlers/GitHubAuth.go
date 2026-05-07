package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"os"
	"stvCms/internal/middleware"
	"stvCms/internal/services"

	"github.com/labstack/echo/v4"
)

type GitHubAuthHandler struct {
	service services.IAuthService
}

func NewGitHubAuthHandler(service services.IAuthService) *GitHubAuthHandler {
	return &GitHubAuthHandler{service: service}
}

type GitHubLoginResponse struct {
	Token string         `json:"token"`
	User  UserProfile    `json:"user"`
}

type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type githubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubRedirect redirects the user to GitHub's authorize page
func (h *GitHubAuthHandler) GitHubRedirect(c echo.Context) error {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	if clientID == "" {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "GitHub OAuth not configured"})
	}

	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&scope=user:email",
		clientID,
	)
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GitHubCallback handles the callback from GitHub
func (h *GitHubAuthHandler) GitHubCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "missing code parameter"})
	}

	// Exchange code for access token
	token, err := exchangeGitHubCode(code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to exchange code: " + err.Error()})
	}

	// Get user info from GitHub
	userInfo, err := getGitHubUserInfo(token.AccessToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get user info: " + err.Error()})
	}

	// Use login as name if name is empty
	name := userInfo.Name
	if name == "" {
		name = userInfo.Login
	}

	// Sync user with database
	githubID := fmt.Sprintf("%d", userInfo.ID)
	dbUser, err := h.service.SyncUserWithGitHub(userInfo.Email, name, userInfo.AvatarURL, githubID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to sync user"})
	}

	// Generate JWT
	jwtToken, err := middleware.GenerateToken(fmt.Sprintf("%d", dbUser.ID), dbUser.Email, dbUser.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to generate token"})
	}

	// Set auth cookie and redirect to frontend with token
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	// Redirect to frontend with token as query param for the callback page
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	redirectURL := fmt.Sprintf("%s/auth/github/callback?token=%s", frontendURL, jwtToken)
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func exchangeGitHubCode(code string) (*githubTokenResponse, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	data := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub returned status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp githubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access token received")
	}

	return &tokenResp, nil
}

func getGitHubUserInfo(accessToken string) (*githubUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var info githubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// If email is empty, try to get it from the emails endpoint
	if info.Email == "" {
		email, err := getGitHubUserEmail(accessToken)
		if err == nil && email != "" {
			info.Email = email
		}
	}

	return &info, nil
}

func getGitHubUserEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", nil
}