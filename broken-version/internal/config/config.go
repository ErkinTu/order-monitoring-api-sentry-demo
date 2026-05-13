package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort                string
	AppName                string
	LogLevel               string
	DatabaseDSN            string
	SentryDSN              string
	SentryEnvironment      string
	SentryTracesSampleRate float64
	ElasticEnabled         bool
	ElasticCloudID         string
	ElasticURL             string
	ElasticAPIKey          string
	ElasticUsername        string
	ElasticPassword        string
	ElasticIndexPrefix     string
	ElasticQueueSize       int
	ElasticTimeoutSeconds  int
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	return Config{
		AppPort:                getEnv("APP_PORT", "8080"),
		AppName:                getEnv("APP_NAME", "order-monitoring-api-broken"),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		DatabaseDSN:            getEnv("DATABASE_DSN", "host=localhost user=app password=app dbname=orders port=5432 sslmode=disable TimeZone=Asia/Bishkek"),
		SentryDSN:              getEnv("SENTRY_DSN", ""),
		SentryEnvironment:      getEnv("SENTRY_ENVIRONMENT", "development"),
		SentryTracesSampleRate: getEnvAsFloat("SENTRY_TRACES_SAMPLE_RATE", 1.0),
		ElasticEnabled:         getEnvAsBool("ELASTIC_ENABLED", false),
		ElasticCloudID:         getEnv("ELASTIC_CLOUD_ID", ""),
		ElasticURL:             getEnv("ELASTIC_URL", ""),
		ElasticAPIKey:          getEnv("ELASTIC_API_KEY", ""),
		ElasticUsername:        getEnv("ELASTIC_USERNAME", ""),
		ElasticPassword:        getEnv("ELASTIC_PASSWORD", ""),
		ElasticIndexPrefix:     getEnv("ELASTIC_INDEX_PREFIX", "order-monitoring-logs"),
		ElasticQueueSize:       getEnvAsInt("ELASTIC_QUEUE_SIZE", 2000),
		ElasticTimeoutSeconds:  getEnvAsInt("ELASTIC_TIMEOUT_SECONDS", 3),
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

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvAsBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
