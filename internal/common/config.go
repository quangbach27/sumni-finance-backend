package common

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	port               string
	appEnv             string
	corsAllowedOrigins []string

	dbUsernane string
	dbPassword string
	dbHost     string
	dbName     string
	dbURL      string

	keycloakRealm      string
	keycloakApiBaseUrl string

	bankLookupBaseUrl string
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
		port:               getEnv("PORT", "4000"),
		appEnv:             getEnv("APP_ENV", "dev"),
		corsAllowedOrigins: strings.Split(corsOrigins, ";"),

		dbUsernane: dbUsername,
		dbPassword: dbPassword,
		dbHost:     dbHost,
		dbName:     dbName,
		dbURL:      dbURL,

		keycloakRealm:      getEnv("KEYCLOAK_REALM", "sumni-finance"),
		keycloakApiBaseUrl: getEnv("KEYCLOAK_API_BASED_URL", "http://localhost:8080"),
		bankLookupBaseUrl:  getEnv("BANK_LOOKUP_BASE_URL", "https://api.vietqr.io"),
	}
}

func (c *Config) Port() string {
	return c.port
}

func (c *Config) AppEnv() string {
	return c.appEnv
}

func (c *Config) CorsAllowedOrigins() []string {
	return c.corsAllowedOrigins
}

func (c *Config) DbUsername() string {
	return c.dbUsernane
}

func (c *Config) DbPassword() string {
	return c.dbPassword
}

func (c *Config) DbHost() string {
	return c.dbHost
}

func (c *Config) DbName() string {
	return c.dbName
}

func (c *Config) DbURL() string {
	return c.dbURL
}

func (c *Config) KeycloakRealm() string {
	return c.keycloakRealm
}

func (c *Config) KeycloakApiBaseUrl() string {
	return c.keycloakApiBaseUrl
}

func (c *Config) BankLookupBaseUrl() string {
	return c.bankLookupBaseUrl
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	slog.Warn(fmt.Sprintf("Environment variable %s is not set. Using fallback value: %s", key, fallback))

	return fallback
}
