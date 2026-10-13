package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// CORS allows the given origins. If the list is empty, all are accepted:
// the request Origin is echoed back, because "*" is not valid together with AllowCredentials.
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
