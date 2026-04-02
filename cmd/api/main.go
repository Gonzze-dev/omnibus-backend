package main

import (
	"log"

	"github.com/labstack/echo/v4"

	"tesina/backend/internal/app"
	"tesina/backend/internal/config"
	"tesina/backend/internal/database"
	"tesina/backend/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.OpenPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	application := app.New(cfg, db)

	e := echo.New()
	router.Register(e, application, cfg)

	log.Println("Server starting on", cfg.ListenAddr)
	if err := e.Start(cfg.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
