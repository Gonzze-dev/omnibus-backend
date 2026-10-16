package middleware

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func Logging() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			// A returned error is written later by Echo's error handler, so the
			// response still says 200: take the status from the error instead.
			status := c.Response().Status
			if err != nil {
				status = http.StatusInternalServerError
				var he *echo.HTTPError
				if errors.As(err, &he) {
					status = he.Code
				}
			}

			log.Printf("%s %s %d %v",
				c.Request().Method,
				c.Request().URL.Path,
				status,
				time.Since(start),
			)

			return err
		}
	}
}
