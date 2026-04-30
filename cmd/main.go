package cmd

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

	svc, err := internal.New(
		ctx,
		dbPgx,
	)
	if err != nil {
		panic(err)
	}

	if err := svc.Run(ctx, ":8080"); err != nil {
		panic(err)
	}
}
