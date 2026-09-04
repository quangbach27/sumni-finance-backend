package common

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type AppConfig struct {
	Port               string
	Env                string
	CorsAllowedOrigins []string
	RateLimitRPS       float64
}

type DBConfig struct {
	Username string
	Password string
	Host     string
	Name     string
	URL      string
}

type Config struct {
	App *AppConfig
	DB  *DBConfig
}

func NewConfig() *Config {
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	dbUsername := getEnv("DB_USERNAME", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbHost := getEnv("DB_HOST", "localhost")
	dbName := getEnv("DB_NAME", "sumni-finance")
	dbPort := getEnv("DB_PORT", "5432")
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUsername, dbPassword, dbHost, dbPort, dbName)

	return &Config{
		App: &AppConfig{
			Port:               getEnv("PORT", "4000"),
			Env:                getEnv("APP_ENV", "dev"),
			CorsAllowedOrigins: parseOrigins(corsOrigins),
			RateLimitRPS:       getEnvFloat("RATE_LIMIT_RPS", 20.0),
		},
		DB: &DBConfig{
			Username: dbUsername,
			Password: dbPassword,
			Host:     dbHost,
			Name:     dbName,
			URL:      dbURL,
		},
	}
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	slog.Debug(fmt.Sprintf("Environment variable %s is not set. Using fallback value: %s", key, fallback))
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		slog.Debug(fmt.Sprintf("Environment variable %s is not set. Using fallback value: %v", key, fallback))
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 {
		slog.Warn("invalid environment variable value, using fallback", "key", key, "value", value, "fallback", fallback)
		return fallback
	}

	return parsed
}

func parseOrigins(raw string) []string {
	parts := strings.Split(raw, ";")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
