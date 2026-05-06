package app

import (
	"log"
	"net/http"

	"gorm.io/gorm"

	"tesina/backend/internal/config"
	"tesina/backend/internal/handler"
	"tesina/backend/internal/mail"
	"tesina/backend/internal/repository"
	"tesina/backend/internal/service"
	realtimeHubMethods  "tesina/backend/internal/realtime"
	"tesina/backend/pkg/realtime"
)

type App struct {
	BusTicket        *handler.BusTicketHandler
	Notification     *handler.NotificationHandler
	Auth             *handler.AuthHandler
	PasswordRecovery *handler.PasswordRecoveryHandler
	User             *handler.UserHandler
	Admin            *handler.AdminHandler
	SuperAdmin       *handler.SuperAdminHandler
	Bus              *handler.BusHandler
}

func New(cfg config.Config, db *gorm.DB) *App {
	cityRepo := repository.NewCityRepository(db)
	busTerminalRepo := repository.NewBusTerminalRepository(db)
	awaitedTripRepo := repository.NewAwaitedTripRepository(db)
	platformRepo := repository.NewPlatformRepository(db)
	userRepo := repository.NewUserRepository(db)
	rolRepo := repository.NewRolRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	userTerminalRepo := repository.NewUserTerminalRepository(db)

	notificationRepo := repository.NewNotificationRepository(db)

	httpClient := &http.Client{
		Timeout: cfg.HTTPClientTimeout,
	}
	BusTicketSvc := service.NewBusTicketService(httpClient, cfg.ExternalTerminalUpstreamURL)

	var smtpMailer *mail.Mailer
	if cfg.SMTPHost != "" {
		m, err := mail.New(mail.Config{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			User:     cfg.SMTPUser,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
		})
		if err != nil {
			log.Printf("password recovery: SMTP mailer disabled: %v", err)
		} else {
			smtpMailer = m
		}
	}

	signalRClient := realtime.NewClient(cfg.RealtimeURL, cfg.RealtimeAPIKey)
	notificationSvc := service.NewNotificationService(platformRepo, userTerminalRepo, busTerminalRepo, notificationRepo, awaitedTripRepo, signalRClient, realtimeHubMethods.DefaultRealtimeHubMethods(), BusTicketSvc, smtpMailer, cfg.MailSiteName)

	authSvc := service.NewAuthService(userRepo, rolRepo, refreshTokenRepo, cfg.JWTSecret)

	recoverySvc := service.NewPasswordRecoveryService(
		userRepo,
		refreshTokenRepo,
		smtpMailer,
		cfg.PasswordResetJWTSecret,
		cfg.FrontEndBaseLink,
		cfg.MailSiteName,
	)

	busSvc := service.NewBusService(busTerminalRepo, awaitedTripRepo, BusTicketSvc)

	userSvc := service.NewUserService(userRepo, refreshTokenRepo, busTerminalRepo, userTerminalRepo)

	adminSvc := service.NewAdminService(cityRepo, platformRepo, busTerminalRepo, userRepo, rolRepo, userTerminalRepo)
	superAdminSvc := service.NewSuperAdminService(cityRepo, busTerminalRepo, userRepo, rolRepo, userTerminalRepo, BusTicketSvc)

	return &App{
		BusTicket:        handler.NewBusTicketHandler(BusTicketSvc),
		Bus:              handler.NewBusHandler(busSvc),
		Notification:     handler.NewNotificationHandler(notificationSvc),
		Auth:             handler.NewAuthHandler(authSvc),
		PasswordRecovery: handler.NewPasswordRecoveryHandler(recoverySvc),
		User:             handler.NewUserHandler(userSvc),
		Admin:            handler.NewAdminHandler(adminSvc),
		SuperAdmin:       handler.NewSuperAdminHandler(superAdminSvc),
	}
}
