package monitoring

import (
	"log"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"

	"github.com/example/order-monitoring-api-fixed/internal/config"
)

func InitSentry(cfg config.Config) error {
	if cfg.SentryDSN == "" {
		log.Println("SENTRY_DSN is empty: Sentry events will not be delivered")
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

	log.Println("Sentry initialized")
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
