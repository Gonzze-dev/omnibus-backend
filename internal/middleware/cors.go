package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func CORS() echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderContentType},
	})
}
