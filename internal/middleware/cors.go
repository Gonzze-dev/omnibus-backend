package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// CORS permite los orígenes indicados. Si la lista está vacía se aceptan todos:
// se refleja el Origin de la petición, porque "*" no es válido junto a AllowCredentials.
func CORS(allowedOrigins []string) echo.MiddlewareFunc {
	cfg := echomw.CORSConfig{
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	}
	if len(allowedOrigins) == 0 {
		cfg.AllowOriginFunc = func(string) (bool, error) { return true, nil }
	} else {
		cfg.AllowOrigins = allowedOrigins
	}
	return echomw.CORSWithConfig(cfg)
}
