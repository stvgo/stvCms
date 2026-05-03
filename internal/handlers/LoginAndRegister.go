package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"stvCms/internal/middleware"
	"stvCms/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
)

type LoginAndRegisterHandler struct {
	service     services.ILoginAndRegisterService
	authService services.IAuthService
}

func NewLoginAndRegisterHandler(authService services.IAuthService) *LoginAndRegisterHandler {
	return &LoginAndRegisterHandler{
		service:     services.NewLoginAndRegisterService(),
		authService: authService,
	}
}

func (h *LoginAndRegisterHandler) Home(c echo.Context) error {
	tmpl, err := template.ParseFiles("././public/templates/index.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error loading template")
	}

	err = tmpl.Execute(c.Response().Writer, nil)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return nil
}

func (h *LoginAndRegisterHandler) SignInWithProvider(c echo.Context) error {
	provider := c.Param("provider")
	q := c.Request().URL.Query()
	q.Add("provider", provider)
	c.Request().URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Response().Writer, c.Request())

	return nil
}

func (h *LoginAndRegisterHandler) CallbackHandler(c echo.Context) error {
	provider := c.Param("provider")
	q := c.Request().URL.Query()
	q.Add("provider", provider)
	c.Request().URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(c.Response().Writer, c.Request())
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Sync user with database
	dbUser, err := h.authService.SyncUser(user.Email, user.Name, user.AvatarURL, user.UserID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error syncing user")
	}

	// Generate JWT and set auth cookie
	token, err := middleware.GenerateToken(fmt.Sprintf("%d", dbUser.ID), dbUser.Email)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error generating token")
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

	// Guardar sesión
	session, _ := gothic.Store.Get(c.Request(), "user-session")
	session.Values["user_id"] = user.UserID
	session.Values["email"] = user.Email
	session.Values["name"] = user.Name
	session.Values["avatar"] = user.AvatarURL
	err = session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error saving session")
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/auth/success")
}

func (h *LoginAndRegisterHandler) Success(c echo.Context) error {
	session, _ := gothic.Store.Get(c.Request(), "user-session")

	email := ""
	name := ""
	avatar := ""

	if val, ok := session.Values["email"].(string); ok {
		email = val
	}
	if val, ok := session.Values["name"].(string); ok {
		name = val
	}
	if val, ok := session.Values["avatar"].(string); ok {
		avatar = val
	}

	html := fmt.Sprintf(`
		<div style="
			font-family: Arial, sans-serif;
			background-color: #f0f0f0;
			display: flex;
			justify-content: center;
			align-items: center;
			height: 100vh;
		">
			<div style="
				background-color: #fff;
				padding: 40px;
				border-radius: 8px;
				box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
				text-align: center;
			">
				<img src="%s" alt="avatar" style="border-radius: 50%%; width: 100px; height: 100px; margin-bottom: 20px;">
				<h1 style="color: #333; margin-bottom: 10px;">Welcome, %s!</h1>
				<p style="color: #666;">%s</p>
			</div>
		</div>
	`, avatar, name, email)

	return c.HTML(http.StatusOK, html)
}
