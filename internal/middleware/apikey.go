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

// CameraAPIKeyOrRole accepts either a valid X-API-Key (camera / edge devices) or a
// Bearer JWT whose role is one of allowedRoles (e.g. an admin notifying manually
// when a camera fails). If X-API-Key is present it takes precedence; with a JWT,
// user_id and role are set in the context like Auth does.
func CameraAPIKeyOrRole(expectedAPIKey, jwtSecret string, allowedRoles ...string) echo.MiddlewareFunc {
	apiKey := CameraAPIKey(expectedAPIKey)
	jwtAuth := func(next echo.HandlerFunc) echo.HandlerFunc {
		return Auth(jwtSecret)(RequireRole(allowedRoles...)(next))
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		withAPIKey := apiKey(next)
		withJWT := jwtAuth(next)
		return func(c echo.Context) error {
			if c.Request().Header.Get("X-API-Key") != "" {
				return withAPIKey(c)
			}
			if c.Request().Header.Get("Authorization") != "" {
				return withJWT(c)
			}
			return echo.NewHTTPError(http.StatusUnauthorized, "missing api key or authorization header")
		}
	}
}
