package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultDatabaseURL                 = "host=localhost user=postgres password=1234 dbname=omnibus-terminal port=5432 sslmode=disable"
	defaultJWTSecret                   = "default-secret-change-me"
	defaultPasswordResetJWTSecret      = "default-password-reset-secret-change-me"
	defaultExternalTerminalUpstreamURL = "http://localhost:4990"
	defaultRealtimeURL                 = "http://localhost:4988/realtime"
	defaultHTTPClientTimeout           = 10 * time.Second
	defaultListenAddr                  = ":4989"
	defaultCameraNotificationAPIKey    = "DEFAULT_API_KEY"
	defaultMailSiteName                = "Omnibus"
	defaultSMTPPort                    = 587
	defaultFrontEndBaseLink            = "http://localhost:4200/"
)

type Config struct {
	DatabaseURL                 string
	JWTSecret                   string
	PasswordResetJWTSecret      string
	FrontEndBaseLink            string
	MailSiteName                string
	SMTPHost                    string
	SMTPPort                    int
	SMTPUser                    string
	SMTPPassword                string
	SMTPFrom                    string
	ExternalTerminalUpstreamURL string
	RealtimeURL                 string
	HTTPClientTimeout           time.Duration
	ListenAddr                  string
	CameraNotificationAPIKey    string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("aviso: no se cargó .env, se usan solo variables del sistema:", err)
	}

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

	passwordResetJWTSecret := os.Getenv("PASSWORD_RESET_JWT_SECRET")
	if passwordResetJWTSecret == "" {
		passwordResetJWTSecret = defaultPasswordResetJWTSecret
	}

	frontEndBaseLink := os.Getenv("FRONT_END_BASE_LINK")
	if frontEndBaseLink == "" {
		frontEndBaseLink = defaultFrontEndBaseLink
	}

	mailSiteName := os.Getenv("MAIL_SITE_NAME")
	if mailSiteName == "" {
		mailSiteName = defaultMailSiteName
	}

	smtpPort := defaultSMTPPort
	if p := os.Getenv("SMTP_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			smtpPort = n
		}
	}

	return Config{
		DatabaseURL:                 dsn,
		JWTSecret:                   jwtSecret,
		PasswordResetJWTSecret:      passwordResetJWTSecret,
		FrontEndBaseLink:            frontEndBaseLink,
		MailSiteName:                mailSiteName,
		SMTPHost:                    os.Getenv("SMTP_HOST"),
		SMTPPort:                    smtpPort,
		SMTPUser:                    os.Getenv("SMTP_USER"),
		SMTPPassword:                os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:                    os.Getenv("SMTP_FROM"),
		ExternalTerminalUpstreamURL: externalTerminalUpstreamURL,
		RealtimeURL:                 realtimeURL,
		HTTPClientTimeout:           defaultHTTPClientTimeout,
		ListenAddr:                  listenAddr,
		CameraNotificationAPIKey:    cameraAPIKey,
	}
}
