package monitoring

import (
	"log/slog"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"github.com/example/order-monitoring-api-broken/internal/config"
)

func InitSentry(cfg config.Config) error {
	if cfg.SentryDSN == "" {
		slog.Info("sentry is disabled because SENTRY_DSN is empty")
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.SentryEnvironment,
		EnableTracing:    true,
		TracesSampleRate: cfg.SentryTracesSampleRate,
	})
	if err != nil {
		return err
	}

	slog.Info("sentry initialized")
	return nil
}

func CaptureException(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
		return
	}

	sentry.CaptureException(err)
}

func CaptureMessage(c *gin.Context, message string) {
	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.CaptureMessage(message)
		return
	}

	sentry.CaptureMessage(message)
}
