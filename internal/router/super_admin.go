package router

import (
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/app"
	"tesina/backend/internal/middleware"
	"tesina/backend/internal/roles"
)

func registerSuperAdmin(e *echo.Echo, a *app.App, jwtSecret string) {
	superAdmin := e.Group("/api/super", middleware.Auth(jwtSecret), middleware.RequireRole(roles.SuperAdmin))
	superAdmin.GET("/terminals", a.SuperAdmin.ListTerminals)
	superAdmin.GET("/terminals/:uuid", a.SuperAdmin.GetTerminal)
	superAdmin.POST("/terminals", a.SuperAdmin.CreateTerminal)
	superAdmin.PUT("/terminals/:uuid", a.SuperAdmin.UpdateTerminal)
	superAdmin.DELETE("/terminals/:uuid", a.SuperAdmin.DeleteTerminal)
	superAdmin.POST("/users/promote-super", a.SuperAdmin.PromoteToSuper)
	superAdmin.POST("/users/demote-super", a.SuperAdmin.DemoteSuper)
}
