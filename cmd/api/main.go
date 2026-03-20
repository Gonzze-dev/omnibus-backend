package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tesina/backend/internal/handler"
	"tesina/backend/internal/middleware"
	"tesina/backend/internal/repository"
	"tesina/backend/internal/service"
	"tesina/backend/pkg/realtime"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=1234 dbname=omnibus-terminal port=5432 sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-me"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database")

	// Repositories
	cityRepo := repository.NewCityRepository(db)
	busTerminalRepo := repository.NewBusTerminalRepository(db)
	platformRepo := repository.NewPlatformRepository(db)
	userRepo := repository.NewUserRepository(db)
	rolRepo := repository.NewRolRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	userTerminalRepo := repository.NewUserTerminalRepository(db)

	// Existing services
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	upstreamURL := "http://localhost:4990"
	pasajeSvc := service.NewPasajeService(httpClient, upstreamURL)

	signalRClient := realtime.NewClient("http://localhost:4988/realtime")
	notificationSvc := service.NewNotificationService(platformRepo, signalRClient)

	// Auth & user services
	authSvc := service.NewAuthService(userRepo, rolRepo, refreshTokenRepo, jwtSecret)
	userSvc := service.NewUserService(userRepo, refreshTokenRepo)

	// Admin & super admin services
	adminSvc := service.NewAdminService(cityRepo, platformRepo, busTerminalRepo, userRepo, rolRepo, userTerminalRepo)
	superAdminSvc := service.NewSuperAdminService(cityRepo, busTerminalRepo, userRepo, rolRepo, userTerminalRepo)

	// Handlers
	pasajeHandler := handler.NewPasajeHandler(pasajeSvc)
	notificationHandler := handler.NewNotificationHandler(notificationSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	adminHandler := handler.NewAdminHandler(adminSvc)
	superAdminHandler := handler.NewSuperAdminHandler(superAdminSvc)

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logging())

	// Public routes
	e.GET("/health", handler.HealthHandler)
	e.GET("/pasajes/:ticket_string", pasajeHandler.GetPasaje)
	e.POST("/notify_passengers", notificationHandler.NotifyPassengers)

	// Auth routes (public)
	auth := e.Group("/api/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.RefreshToken)
	auth.POST("/logout", authHandler.Logout)

	// User routes (authenticated)
	users := e.Group("/api/users", middleware.Auth())
	users.GET("/me", userHandler.GetProfile)
	users.PUT("/me", userHandler.UpdateProfile)
	users.DELETE("/me", userHandler.DeleteAccount)

	// Admin routes (authenticated + admin or super_admin role)
	admin := e.Group("/api/admin", middleware.Auth(), middleware.RequireRole("admin", "super_admin"))
	admin.GET("/cities", adminHandler.ListCities)
	admin.GET("/cities/:postal_code", adminHandler.GetCity)
	admin.POST("/cities", adminHandler.CreateCity)
	admin.PUT("/cities/:postal_code", adminHandler.UpdateCity)
	admin.DELETE("/cities/:postal_code", adminHandler.DeleteCity)
	admin.GET("/platforms", adminHandler.ListPlatforms)
	admin.GET("/platforms/:code", adminHandler.GetPlatform)
	admin.POST("/platforms", adminHandler.CreatePlatform)
	admin.PUT("/platforms/:code", adminHandler.UpdatePlatform)
	admin.DELETE("/platforms/:code", adminHandler.DeletePlatform)
	admin.POST("/users/promote", adminHandler.PromoteToAdmin)
	admin.POST("/users/demote", adminHandler.DemoteAdmin)

	// Super admin routes (authenticated + super_admin role)
	// Cities, platforms, and promote/demote admin are handled by the admin group above.
	superAdmin := e.Group("/api/super", middleware.Auth(), middleware.RequireRole("super_admin"))
	superAdmin.GET("/terminals", superAdminHandler.ListTerminals)
	superAdmin.GET("/terminals/:uuid", superAdminHandler.GetTerminal)
	superAdmin.POST("/terminals", superAdminHandler.CreateTerminal)
	superAdmin.PUT("/terminals/:uuid", superAdminHandler.UpdateTerminal)
	superAdmin.DELETE("/terminals/:uuid", superAdminHandler.DeleteTerminal)
	superAdmin.POST("/users/promote-super", superAdminHandler.PromoteToSuper)
	superAdmin.POST("/users/demote-super", superAdminHandler.DemoteSuper)

	log.Println("Server starting on :4989")
	if err := e.Start(":4989"); err != nil {
		log.Fatal(err)
	}
}
