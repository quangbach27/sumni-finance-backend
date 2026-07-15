package common

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type AppConfig struct {
	Port               string
	Env                string
	CorsAllowedOrigins []string
	BankLookupBaseUrl  string
}

type DBConfig struct {
	Username string
	Password string
	Host     string
	Name     string
	URL      string
}

type KeycloakConfig struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
}

type AppAuthConfig struct {
	RedirectURL        string
	PostLoginRedirect  string
	PostLogoutRedirect string
}

// Master Config wrapping the sub-configs
type Config struct {
	App      AppConfig
	DB       DBConfig
	Keycloak KeycloakConfig
	Auth     AppAuthConfig
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
		App: AppConfig{
			Port:               getEnv("PORT", "4000"),
			Env:                getEnv("APP_ENV", "dev"),
			CorsAllowedOrigins: strings.Split(corsOrigins, ";"),
			BankLookupBaseUrl:  getEnv("BANK_LOOKUP_BASE_URL", "https://api.vietqr.io"),
		},
		DB: DBConfig{
			Username: dbUsername,
			Password: dbPassword,
			Host:     dbHost,
			Name:     dbName,
			URL:      dbURL,
		},
		Keycloak: KeycloakConfig{
			BaseURL:      getEnv("KEYCLOAK_BASE_URL", "http://localhost:8080"),
			Realm:        getEnv("KEYCLOAK_REALM", "sumni-finance"),
			ClientID:     getEnv("KEYCLOAK_CLIENT_ID", "sumni-finance-backend"),
			ClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", ""),
		},
		Auth: AppAuthConfig{
			RedirectURL:        getEnv("APP_AUTH_REDIRECT_URL", "http://localhost:4000/auth/callback"),
			PostLoginRedirect:  getEnv("APP_AUTH_POST_LOGIN_URL", "/v1/treasury/fund-source"),
			PostLogoutRedirect: getEnv("APP_AUTH_POST_LOGOUT_URL", "/auth/login"),
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
