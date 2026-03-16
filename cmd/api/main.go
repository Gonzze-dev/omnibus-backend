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

	_ = repository.NewCityRepository(db)
	_ = repository.NewBusTerminalRepository(db)
	platformRepo := repository.NewPlatformRepository(db)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	upstreamURL := "http://localhost:4990"
	pasajeSvc := service.NewPasajeService(httpClient, upstreamURL)

	signalRClient := realtime.NewClient("http://localhost:4988/realtime")
	notificationSvc := service.NewNotificationService(platformRepo, signalRClient)

	pasajeHandler := handler.NewPasajeHandler(pasajeSvc)
	notificationHandler := handler.NewNotificationHandler(notificationSvc)

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logging())

	e.GET("/health", handler.HealthHandler)
	e.GET("/pasajes/:ticket_string", pasajeHandler.GetPasaje)
	e.POST("/notify_passengers", notificationHandler.NotifyPassengers)

	log.Println("Server starting on :4989")
	if err := e.Start(":4989"); err != nil {
		log.Fatal(err)
	}
}
