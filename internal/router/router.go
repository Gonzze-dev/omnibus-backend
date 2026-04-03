// Package router registers HTTP routes on Echo.
package router

import (
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/app"
	"tesina/backend/internal/config"
	"tesina/backend/internal/middleware"
)

func Register(e *echo.Echo, a *app.App, cfg config.Config) {
	e.Use(middleware.CORS())
	e.Use(middleware.Logging())

	registerPublic(e, a, cfg)
	registerUser(e, a, cfg.JWTSecret)
	registerAdmin(e, a, cfg.JWTSecret)
	registerSuperAdmin(e, a, cfg.JWTSecret)
}
