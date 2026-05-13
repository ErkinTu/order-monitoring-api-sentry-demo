package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/example/order-monitoring-api-fixed/internal/config"
	"github.com/example/order-monitoring-api-fixed/internal/database"
	"github.com/example/order-monitoring-api-fixed/internal/logging"
	"github.com/example/order-monitoring-api-fixed/internal/monitoring"
	"github.com/example/order-monitoring-api-fixed/internal/router"
)

func main() {
	cfg := config.Load()

	_, closeLogs, err := logging.Setup(logging.Options{
		AppName:               cfg.AppName,
		Environment:           cfg.SentryEnvironment,
		Level:                 cfg.LogLevel,
		ElasticEnabled:        cfg.ElasticEnabled,
		ElasticCloudID:        cfg.ElasticCloudID,
		ElasticURL:            cfg.ElasticURL,
		ElasticAPIKey:         cfg.ElasticAPIKey,
		ElasticUsername:       cfg.ElasticUsername,
		ElasticPassword:       cfg.ElasticPassword,
		ElasticIndexPrefix:    cfg.ElasticIndexPrefix,
		ElasticQueueSize:      cfg.ElasticQueueSize,
		ElasticTimeoutSeconds: cfg.ElasticTimeoutSeconds,
	})
	if err != nil {
		log.Fatalf("logger initialization failed: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := closeLogs(ctx); err != nil {
			slog.Warn("failed to flush elastic log queue", "error", err)
		}
	}()

	if err := monitoring.InitSentry(cfg); err != nil {
		slog.Error("sentry initialization failed", "error", err)
	}
	defer sentry.Flush(2 * time.Second)

	db, err := database.ConnectWithRetry(cfg)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		return
	}

	if err := database.AutoMigrate(db); err != nil {
		slog.Error("database migration failed", "error", err)
		return
	}

	if err := database.SeedProducts(db); err != nil {
		slog.Error("database seed failed", "error", err)
		return
	}

	r := router.New(db)

	slog.Info("server is running", "port", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		slog.Error("server stopped", "error", err)
	}
}
