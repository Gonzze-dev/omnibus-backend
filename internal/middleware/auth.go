package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

func Auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format")
			}

			secret := os.Getenv("JWT_SECRET")
			if secret == "" {
				secret = "default-secret-change-me"
			}

			claims := &JWTClaims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "unexpected signing method")
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			c.Set("user_id", claims.UserID)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}
