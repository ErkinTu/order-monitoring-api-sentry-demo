package main

import (
	"log"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/example/order-monitoring-api-fixed/internal/config"
	"github.com/example/order-monitoring-api-fixed/internal/database"
	"github.com/example/order-monitoring-api-fixed/internal/monitoring"
	"github.com/example/order-monitoring-api-fixed/internal/router"
)

func main() {
	cfg := config.Load()

	if err := monitoring.InitSentry(cfg); err != nil {
		log.Printf("sentry initialization failed: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	db, err := database.ConnectWithRetry(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	if err := database.SeedProducts(db); err != nil {
		log.Fatalf("database seed failed: %v", err)
	}

	r := router.New(db)

	log.Printf("server is running on port %s", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
