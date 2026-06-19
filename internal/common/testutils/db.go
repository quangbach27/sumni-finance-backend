package testutils

import (
	"context"
	"io/fs"

	"sumni-finance-backend/internal/common"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(moduleName string, embedFS fs.FS, migrationsDir string) {
	ctx := context.Background()
	config := common.NewConfig()

	pool, err := pgxpool.New(ctx, config.DbURL())
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	if err := common.MigrateDatabaseUp(ctx, moduleName, pool, embedFS, migrationsDir); err != nil {
		panic(err)
	}
}

func NewDB(ctx context.Context) (*pgxpool.Pool, func()) {
	config := common.NewConfig()

	pool, err := pgxpool.New(ctx, config.DbURL())
	if err != nil {
		panic(err)
	}

	return pool, func() { pool.Close() }
}
