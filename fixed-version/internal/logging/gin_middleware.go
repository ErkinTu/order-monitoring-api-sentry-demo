package logging

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		fields := []any{
			"http.method", c.Request.Method,
			"url.path", path,
			"http.status_code", status,
			"event.duration_ms", latency.Milliseconds(),
			"client.address", c.ClientIP(),
			"user_agent.original", c.Request.UserAgent(),
		}

		if len(c.Errors) > 0 || status >= 500 {
			if len(c.Errors) > 0 {
				fields = append(fields, "errors", c.Errors.String())
			}
			slog.Error("http request completed with server error", fields...)
			return
		}

		if status >= 400 {
			slog.Warn("http request completed with client error", fields...)
			return
		}

		slog.Info("http request completed", fields...)
	}
}
