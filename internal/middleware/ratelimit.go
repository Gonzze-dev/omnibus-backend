package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// RateLimit allows each client IP (c.RealIP()) perMinute requests per minute on
// average, with bursts of up to burst requests; beyond that it responds 429.
// The counters live in memory, so each limiter (and each replica) counts on its own.
func RateLimit(perMinute float64, burst int) echo.MiddlewareFunc {
	return echomw.RateLimiter(echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{
		Rate:      rate.Limit(perMinute / 60),
		Burst:     burst,
		ExpiresIn: 10 * time.Minute,
	}))
}
