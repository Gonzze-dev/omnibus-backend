package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func HealthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func RootRedirectHandler(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/health")
}
