package router

import (
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/app"
	"tesina/backend/internal/handler"
)

func registerPublic(e *echo.Echo, a *app.App) {
	e.GET("/health", handler.HealthHandler)
	e.GET("/bus_tickets/:ticket_string", a.BusTicket.GetBusTicket)
	e.POST("/notify_passengers", a.Notification.NotifyPassengers)

	auth := e.Group("/api/auth")
	auth.POST("/register", a.Auth.Register)
	auth.POST("/login", a.Auth.Login)
	auth.POST("/refresh", a.Auth.RefreshToken)
	auth.POST("/logout", a.Auth.Logout)
}
