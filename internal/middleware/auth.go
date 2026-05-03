package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var secretKey = []byte(os.Getenv("AUTH_SECRET"))

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "stv-cms",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var tokenStr string

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenStr = parts[1]
				}
			}

			if tokenStr == "" {
				cookie, err := c.Cookie("auth_token")
				if err == nil && cookie.Value != "" {
					tokenStr = cookie.Value
				}
			}

			if tokenStr != "" {
				claims, err := ValidateToken(tokenStr)
				if err != nil {
					return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
				}

				c.Set("user_id", claims.UserID)
				c.Set("email", claims.Email)
				return next(c)
			}

			// Fallback para pruebas locales en Insomnia/Postman
			localUserID := c.Request().Header.Get("X-Local-User-ID")
			if localUserID == "123" {
				c.Set("user_id", "123")
				c.Set("email", "test@local.dev")
				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Authorization header or auth_token cookie is required"})
		}
	}
}
