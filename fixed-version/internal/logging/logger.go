package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Options struct {
	AppName               string
	Environment           string
	Level                 string
	ElasticEnabled        bool
	ElasticCloudID        string
	ElasticURL            string
	ElasticAPIKey         string
	ElasticUsername       string
	ElasticPassword       string
	ElasticIndexPrefix    string
	ElasticQueueSize      int
	ElasticTimeoutSeconds int
}

func Setup(opts Options) (*slog.Logger, func(context.Context) error, error) {
	writers := []io.Writer{os.Stdout}
	closeFn := func(context.Context) error { return nil }

	if opts.ElasticEnabled {
		sink, err := newElasticSink(opts)
		if err != nil {
			return nil, nil, err
		}
		writers = append(writers, sink)
		closeFn = sink.Close
	}

	level := parseLevel(opts.Level)
	handler := slog.NewJSONHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler).With(
		"service.name", opts.AppName,
		"deployment.environment", opts.Environment,
	)
	slog.SetDefault(logger)

	// Route standard log package calls through slog with INFO level.
	slog.SetLogLoggerLevel(slog.LevelInfo)

	logger.Info("logger initialized", "log.level", level.String(), "elastic.enabled", opts.ElasticEnabled)
	return logger, closeFn, nil
}

func parseLevel(level string) slog.Level {
	var parsed slog.Level
	if err := parsed.UnmarshalText([]byte(level)); err != nil {
		return slog.LevelInfo
	}
	return parsed
}

func validateElasticOptions(opts Options) error {
	if opts.ElasticCloudID == "" && opts.ElasticURL == "" {
		return fmt.Errorf("ELASTIC_ENABLED=true requires ELASTIC_CLOUD_ID or ELASTIC_URL")
	}
	if opts.ElasticAPIKey == "" && (opts.ElasticUsername == "" || opts.ElasticPassword == "") {
		return fmt.Errorf("set ELASTIC_API_KEY or ELASTIC_USERNAME/ELASTIC_PASSWORD")
	}
	return nil
}
