package config

import (
	"os"
	"time"
)

const (
	defaultDatabaseURL                 = "host=localhost user=postgres password=1234 dbname=omnibus-terminal port=5432 sslmode=disable"
	defaultJWTSecret                   = "default-secret-change-me"
	defaultExternalTerminalUpstreamURL = "http://localhost:4990"
	defaultRealtimeURL                 = "http://localhost:4988/realtime"
	defaultHTTPClientTimeout           = 10 * time.Second
	defaultListenAddr                  = ":4989"
	defaultCameraNotificationAPIKey    = "DEFAULT_API_KEY"
)

type Config struct {
	DatabaseURL                 string
	JWTSecret                   string
	ExternalTerminalUpstreamURL string
	RealtimeURL                 string
	HTTPClientTimeout           time.Duration
	ListenAddr                  string
	CameraNotificationAPIKey    string
}

func Load() Config {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDatabaseURL
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = defaultJWTSecret
	}

	externalTerminalUpstreamURL := os.Getenv("EXTERN_TERMINAL_UPSTREAM_URL")
	if externalTerminalUpstreamURL == "" {
		externalTerminalUpstreamURL = defaultExternalTerminalUpstreamURL
	}

	realtimeURL := os.Getenv("REALTIME_URL")
	if realtimeURL == "" {
		realtimeURL = defaultRealtimeURL
	}

	cameraAPIKey := os.Getenv("CAMERA_NOTIFICATION_API_KEY")
	if cameraAPIKey == "" {
		cameraAPIKey = defaultCameraNotificationAPIKey
	}

	return Config{
		DatabaseURL:                 dsn,
		JWTSecret:                   jwtSecret,
		ExternalTerminalUpstreamURL: externalTerminalUpstreamURL,
		RealtimeURL:                 realtimeURL,
		HTTPClientTimeout:           defaultHTTPClientTimeout,
		ListenAddr:                  listenAddr,
		CameraNotificationAPIKey:    cameraAPIKey,
	}
}
