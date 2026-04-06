package router

import (
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/app"
	"tesina/backend/internal/config"
	"tesina/backend/internal/handler"
	"tesina/backend/internal/middleware"
)

func registerPublic(e *echo.Echo, a *app.App, cfg config.Config) {
	e.GET("/health", handler.HealthHandler)
	e.GET("/bus_tickets/:ticket_string", a.BusTicket.GetBusTicket)
	e.POST("/notify_passengers", a.Notification.NotifyPassengers, middleware.CameraAPIKey(cfg.CameraNotificationAPIKey))
	e.POST("/notify_camera_error", a.Notification.NotifyCameraError, middleware.CameraAPIKey(cfg.CameraNotificationAPIKey))

	auth := e.Group("/api/auth")
	auth.POST("/register", a.Auth.Register)
	auth.POST("/login", a.Auth.Login)
	auth.POST("/refresh", a.Auth.RefreshToken)
	auth.POST("/logout", a.Auth.Logout)
	auth.POST("/forgot-password", a.PasswordRecovery.ForgotPassword)
	auth.POST("/validate-recovery-token", a.PasswordRecovery.ValidateRecoveryToken)
	auth.POST("/reset-password", a.PasswordRecovery.ResetPassword)
}
