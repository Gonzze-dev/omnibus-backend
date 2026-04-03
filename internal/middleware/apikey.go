package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/labstack/echo/v4"
)

// CameraAPIKey validates X-API-Key against the configured secret (e.g. camera / edge devices).
func CameraAPIKey(expected string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			given := c.Request().Header.Get("X-API-Key")
			if len(given) != len(expected) || subtle.ConstantTimeCompare([]byte(given), []byte(expected)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or missing api key")
			}
			return next(c)
		}
	}
}
