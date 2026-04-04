package app

import (
	"net/http"

	"gorm.io/gorm"

	"tesina/backend/internal/config"
	"tesina/backend/internal/handler"
	"tesina/backend/internal/repository"
	"tesina/backend/internal/service"
	"tesina/backend/pkg/realtime"
)

type App struct {
	BusTicket    *handler.BusTicketHandler
	Notification *handler.NotificationHandler
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	Admin        *handler.AdminHandler
	SuperAdmin   *handler.SuperAdminHandler
}

func New(cfg config.Config, db *gorm.DB) *App {
	cityRepo := repository.NewCityRepository(db)
	busTerminalRepo := repository.NewBusTerminalRepository(db)
	platformRepo := repository.NewPlatformRepository(db)
	userRepo := repository.NewUserRepository(db)
	rolRepo := repository.NewRolRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	userTerminalRepo := repository.NewUserTerminalRepository(db)

	httpClient := &http.Client{
		Timeout: cfg.HTTPClientTimeout,
	}
	BusTicketSvc := service.NewBusTicketService(httpClient, cfg.ExternalTerminalUpstreamURL)

	signalRClient := realtime.NewClient(cfg.RealtimeURL)
	notificationSvc := service.NewNotificationService(platformRepo, userTerminalRepo, busTerminalRepo, signalRClient, BusTicketSvc)

	authSvc := service.NewAuthService(userRepo, rolRepo, refreshTokenRepo, cfg.JWTSecret)
	userSvc := service.NewUserService(userRepo, refreshTokenRepo, busTerminalRepo, userTerminalRepo)

	adminSvc := service.NewAdminService(cityRepo, platformRepo, busTerminalRepo, userRepo, rolRepo, userTerminalRepo)
	superAdminSvc := service.NewSuperAdminService(cityRepo, busTerminalRepo, userRepo, rolRepo, userTerminalRepo, BusTicketSvc)

	return &App{
		BusTicket:    handler.NewBusTicketHandler(BusTicketSvc),
		Notification: handler.NewNotificationHandler(notificationSvc),
		Auth:         handler.NewAuthHandler(authSvc),
		User:         handler.NewUserHandler(userSvc),
		Admin:        handler.NewAdminHandler(adminSvc),
		SuperAdmin:   handler.NewSuperAdminHandler(superAdminSvc),
	}
}
