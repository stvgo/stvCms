package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const AdminEmail = "jsvaleriano321@gmail.com"

// AdminMiddleware blocks access unless the authenticated user is the admin.
func AdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			email, ok := c.Get("email").(string)
			if !ok || email != AdminEmail {
				return c.JSON(http.StatusForbidden, echo.Map{
					"error": "acceso denegado: se requiere rol de administrador",
				})
			}
			return next(c)
		}
	}
}

// IsAdmin checks if the authenticated user is the admin.
func IsAdmin(c echo.Context) bool {
	email, ok := c.Get("email").(string)
	return ok && email == AdminEmail
}

// GetUserEmail extracts the email from the context.
func GetUserEmail(c echo.Context) string {
	if email, ok := c.Get("email").(string); ok {
		return email
	}
	return ""
}