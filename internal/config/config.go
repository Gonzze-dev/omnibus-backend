package config

import (
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultDBHost                      = "localhost"
	defaultDBUser                      = "postgres"
	defaultDBPassword                  = "1234"
	defaultDBName                      = "omnibus-terminal"
	defaultDBPort                      = "5432"
	defaultDBSSLMode                   = "disable"
	defaultJWTSecret                   = "default-secret-change-me"
	defaultPasswordResetJWTSecret      = "default-password-reset-secret-change-me"
	defaultExternalTerminalUpstreamURL = "http://localhost:4990"
	defaultRealtimeURL                 = "http://localhost:4988/realtime"
	defaultHTTPClientTimeout           = 10 * time.Second
	defaultListenAddr                  = ":4989"
	defaultCameraNotificationAPIKey    = "DEFAULT_API_KEY"
	defaultRealtimeAPIKey               = "agag2J26sgJ2SAV6ATAJ6aG26sg26JG"
	defaultMailSiteName                = "Omnibus"
	defaultSMTPPort                    = 587
	defaultFrontEndBaseLink            = "http://localhost:4200/"
	defaultOCRURL                      = "http://localhost:8000"
	defaultOCRAPIKey                   = "default-api-key-12345"
	defaultOCRTimeout                  = 30 * time.Second
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
	RealtimeAPIKey              string
	OCRURL                      string
	OCRAPIKey                   string
	OCRTimeout                  time.Duration
	// CORSAllowedOrigins vacío significa que se permiten todos los orígenes.
	CORSAllowedOrigins []string
}

// getEnvTrim lee una variable de entorno sin espacios sobrantes y usa def si está vacía.
func getEnvTrim(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// buildDatabaseURL arma la URL de PostgreSQL a partir de DB_URL (host), DB_USER,
// DB_PASSWORD, DB_NAME, DB_PORT y DB_SSLMODE.
func buildDatabaseURL() string {
	host := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(os.Getenv("DB_URL")), "host="))
	if host == "" {
		host = defaultDBHost
	}
	user := getEnvTrim("DB_USER", defaultDBUser)
	password := getEnvTrim("DB_PASSWORD", defaultDBPassword)
	name := getEnvTrim("DB_NAME", defaultDBName)
	port := getEnvTrim("DB_PORT", defaultDBPort)
	sslmode := getEnvTrim("DB_SSLMODE", defaultDBSSLMode)

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + name,
		RawQuery: url.Values{"sslmode": {sslmode}}.Encode(),
	}
	return u.String()
}

// parseCSV separa una lista por comas descartando espacios y elementos vacíos.
func parseCSV(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("aviso: no se cargó .env, se usan solo variables del sistema:", err)
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	dsn := buildDatabaseURL()

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

	realtimeAPIKey := os.Getenv("REALTIME_API_KEY")
	if realtimeAPIKey == "" {
		realtimeAPIKey = defaultRealtimeAPIKey
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

	ocrURL := os.Getenv("OCR_URL")
	if ocrURL == "" {
		ocrURL = defaultOCRURL
	}

	ocrAPIKey := os.Getenv("OCR_API_KEY")
	if ocrAPIKey == "" {
		ocrAPIKey = defaultOCRAPIKey
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
		RealtimeAPIKey:              realtimeAPIKey,
		OCRURL:                      ocrURL,
		OCRAPIKey:                   ocrAPIKey,
		OCRTimeout:                  defaultOCRTimeout,
		CORSAllowedOrigins:          parseCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}
}
