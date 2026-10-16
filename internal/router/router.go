// Package router registers HTTP routes on Echo.
package router

import (
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/app"
	"tesina/backend/internal/config"
	"tesina/backend/internal/middleware"
)

func Register(e *echo.Echo, a *app.App, cfg config.Config) {
	// The client IP comes in X-Real-IP (set by the frontend nginx), trusted only
	// when the direct peer is on a private network (Railway's internal network).
	e.IPExtractor = echo.ExtractIPFromRealIPHeader()

	e.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	e.Use(middleware.Logging())
	// Second layer behind the frontend nginx limits, looser than those.
	e.Use(middleware.RateLimit(1200, 80))

	registerPublic(e, a, cfg)
	registerUser(e, a, cfg.JWTSecret)
	registerAdmin(e, a, cfg.JWTSecret)
	registerSuperAdmin(e, a, cfg.JWTSecret)
}
