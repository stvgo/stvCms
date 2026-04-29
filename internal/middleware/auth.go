package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func parseCookies(cookieHeader string) map[string]string {
	cookies := make(map[string]string)
	for _, part := range strings.Split(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			cookies[kv[0]] = kv[1]
		}
	}
	return cookies
}

func VerifyAuthToken(cookieHeader, secret string) (jwt.MapClaims, error) {
	cookies := parseCookies(cookieHeader)
	tokenString, ok := cookies["authjs.session-token"]
	if !ok || tokenString == "" {
		return nil, fmt.Errorf("missing session token")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func JWTMiddleware(authSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookieHeader := c.Request().Header.Get("Cookie")
			claims, err := VerifyAuthToken(cookieHeader, authSecret)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
			}

			email, ok := claims["email"].(string)
			if !ok || email == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
			}

			c.Set("userEmail", email)
			return next(c)
		}
	}
}
