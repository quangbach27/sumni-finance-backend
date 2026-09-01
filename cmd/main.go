package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"sumni-finance-backend/internal"
	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/log"
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

	externalSerivce := internal.ExternalService{}

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
