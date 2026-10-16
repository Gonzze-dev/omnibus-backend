package router

import (
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/app"
	"tesina/backend/internal/config"
	"tesina/backend/internal/docs"
	"tesina/backend/internal/handler"
	"tesina/backend/internal/middleware"
	"tesina/backend/internal/roles"
)

func registerPublic(e *echo.Echo, a *app.App, cfg config.Config) {
	e.GET("/", handler.RootRedirectHandler)
	e.GET("/health", handler.HealthHandler)
	e.GET("/docs", docs.UIHandler)
	e.GET("/docs/openapi.yaml", docs.SpecHandler)
	e.GET("/bus_tickets/:ticket_string", a.BusTicket.GetBusTicket)
	e.POST("/notify_passengers", a.Notification.NotifyPassengers,
		middleware.CameraAPIKeyOrRole(cfg.CameraNotificationAPIKey, cfg.JWTSecret, roles.Admin, roles.SuperAdmin))
	e.POST("/notify_camera_error", a.Notification.NotifyCameraError, middleware.CameraAPIKey(cfg.CameraNotificationAPIKey))
	e.GET("/api/notifications", a.Notification.GetNotifications, middleware.Auth(cfg.JWTSecret))

	auth := e.Group("/api/auth")
	// Brute force on passwords and recovery tokens, and mass sending of emails.
	strict := middleware.RateLimit(10, 10)
	auth.POST("/register", a.Auth.Register, strict)
	auth.POST("/login", a.Auth.Login, strict)
	auth.POST("/refresh", a.Auth.RefreshToken)
	auth.POST("/logout", a.Auth.Logout)
	auth.POST("/forgot-password", a.PasswordRecovery.ForgotPassword, strict)
	auth.POST("/validate-recovery-token", a.PasswordRecovery.ValidateRecoveryToken, strict)
	auth.POST("/reset-password", a.PasswordRecovery.ResetPassword, strict)
}
