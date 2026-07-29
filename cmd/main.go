package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sumni-finance-backend/internal"
	"sumni-finance-backend/internal/common/log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Init(slog.LevelInfo)

	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		panic("POSTGRES_URL environment variable is not set")
	}

	dbPgx, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}

	externalSerivce := internal.ExternalService{}

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
