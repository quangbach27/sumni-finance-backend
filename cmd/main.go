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
	"sumni-finance-backend/internal/treasury/adapters/bank/lookup"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Init(slog.LevelInfo)

	config := common.NewConfig()

	dbPgx, err := pgxpool.New(ctx, config.DbURL())
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

	externalSerivce := internal.ExternalService{
		BankLookupProvider: lookup.NewClient(httpClient, config.BankLookupBaseUrl()),
	}

	svc, err := internal.New(
		ctx,
		dbPgx,
		externalSerivce,
	)
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	if err := svc.Run(ctx, ":"+port); err != nil {
		panic(err)
	}
}
