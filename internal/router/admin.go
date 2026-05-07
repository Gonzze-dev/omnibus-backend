package router

import (
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/app"
	"tesina/backend/internal/middleware"
	"tesina/backend/internal/roles"
)

func registerAdmin(e *echo.Echo, a *app.App, jwtSecret string) {
	admin := e.Group("/api/admin", middleware.Auth(jwtSecret), middleware.RequireRole(roles.Admin, roles.SuperAdmin))
	admin.GET("/cities", a.Admin.ListCities)
	admin.GET("/cities/:postal_code", a.Admin.GetCity)
	admin.POST("/cities", a.Admin.CreateCity)
	admin.PUT("/cities/:postal_code", a.Admin.UpdateCity)
	admin.DELETE("/cities/:postal_code", a.Admin.DeleteCity)
	admin.GET("/platforms", a.Admin.ListPlatforms)
	admin.GET("/platforms/:code", a.Admin.GetPlatform)
	admin.POST("/platforms", a.Admin.CreatePlatform)
	admin.PUT("/platforms/:code", a.Admin.UpdatePlatform)
	admin.DELETE("/platforms/:code", a.Admin.DeletePlatform)
	admin.GET("/users/by-email", a.Admin.GetUserByEmail)
	admin.POST("/users/promote", a.Admin.PromoteToAdmin)
	admin.POST("/users/demote", a.Admin.DemoteAdmin)
	admin.GET("/notification-types", a.Notification.ListAdminNotificationTypes)
	admin.POST("/notifications", a.Notification.SendAdminNotification)
	admin.DELETE("/notifications", a.Notification.DeleteNotification)
	admin.POST("/notify-bus-delay", a.Notification.NotifyBusDelay)
}
