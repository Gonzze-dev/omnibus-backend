package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// TODO: replace AllowOrigins with the actual frontend origin (e.g. "http://localhost:3000")
// when a browser client is integrated. Wildcard "*" cannot be used with AllowCredentials in browsers.
func CORS() echo.MiddlewareFunc {
	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	})
}
