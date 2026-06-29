package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sumni-finance-backend/internal"
	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/log"
	"sumni-finance-backend/internal/identity/adapters/keycloak"
	bankLookup "sumni-finance-backend/internal/treasury/adapters/bank/lookup"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Init(slog.LevelInfo)

	config := common.NewConfig()

	dbPgx, err := pgxpool.New(ctx, config.DB.URL)
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   30 * time.Second,
			DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
			ResponseHeaderTimeout: 30 * time.Second,
			MaxIdleConns:          50,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
		},
	}

	keycloakAuth, err := keycloak.NewAuthenticator(ctx, keycloak.AuthenticatorConfig{
		BaseURL:      config.Keycloak.BaseURL,
		Realm:        config.Keycloak.Realm,
		ClientID:     config.Keycloak.ClientID,
		ClientSecret: config.Keycloak.ClientSecret,
		RedirectURL:  config.Auth.RedirectURL,
	})
	if err != nil {
		panic(err)
	}

	externalSerivce := internal.ExternalService{
		BankLookupProvider: bankLookup.NewClient(httpClient, config.App.BankLookupBaseUrl),
		Authenticator:      keycloakAuth,
		SessionManager:     keycloakAuth,
	}

	svc, err := internal.New(
		ctx,
		config,
		dbPgx,
		externalSerivce,
	)
	if err != nil {
		panic(err)
	}

	if err := svc.Run(ctx, ":"+config.App.Port); err != nil {
		panic(err)
	}
}
