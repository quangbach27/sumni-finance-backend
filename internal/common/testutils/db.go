package testutils

import (
	"context"
	"io/fs"
	"os"
	"testing"
	"time"

	"sumni-finance-backend/internal/common"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func RunMigrations(moduleName string, embedFS fs.FS, migrationsDir string) {
	ctx := context.Background()

	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/sumni-finance?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := common.MigrateDatabaseUp(ctx, moduleName, pool, embedFS, migrationsDir); err != nil {
		panic(err)
	}
}

func NewDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/sumni-finance?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)

	config.MaxConns = 30 // enough for concurrent tests
	config.MinConns = 5  // pre-warm connections
	config.MaxConnIdleTime = 30 * time.Second
	config.ConnConfig.ConnectTimeout = 5 * time.Second

	dbPgx, err := pgxpool.NewWithConfig(t.Context(), config)
	require.NoError(t, err)

	return dbPgx
}
