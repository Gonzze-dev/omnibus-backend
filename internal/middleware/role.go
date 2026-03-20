package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get("role").(string)
			if !ok || role == "" {
				return echo.NewHTTPError(http.StatusForbidden, "role not found in context")
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}
