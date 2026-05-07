package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort                string
	DatabaseDSN            string
	SentryDSN              string
	SentryEnvironment      string
	SentryTracesSampleRate float64
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	return Config{
		AppPort:                getEnv("APP_PORT", "8080"),
		DatabaseDSN:            getEnv("DATABASE_DSN", "host=localhost user=app password=app dbname=orders port=5432 sslmode=disable TimeZone=Asia/Bishkek"),
		SentryDSN:              getEnv("SENTRY_DSN", ""),
		SentryEnvironment:      getEnv("SENTRY_ENVIRONMENT", "development"),
		SentryTracesSampleRate: getEnvAsFloat("SENTRY_TRACES_SAMPLE_RATE", 1.0),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}

	return parsed
}
