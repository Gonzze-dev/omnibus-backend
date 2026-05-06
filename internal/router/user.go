package router

import (
	"github.com/labstack/echo/v4"

	"tesina/backend/internal/app"
	"tesina/backend/internal/middleware"
)

func registerUser(e *echo.Echo, a *app.App, jwtSecret string) {
	users := e.Group("/api/users", middleware.Auth(jwtSecret))
	users.GET("/me", a.User.GetProfile)
	users.PUT("/me", a.User.UpdateProfile)
	users.DELETE("/me", a.User.DeleteAccount)
	users.GET("/terminals", a.User.ListTerminals)

	buses := e.Group("/api/buses", middleware.Auth(jwtSecret))
	buses.POST("/join", a.Bus.JoinBus)
}
