package testutils

import (
	"context"
	"io/fs"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"sumni-finance-backend/internal/common"
)

var (
	dbPool     *pgxpool.Pool
	dbPoolOnce sync.Once
)

func RunMigrations(moduleName string, embedFS fs.FS, migrationsDir string) {
	ctx := context.Background()

	pool := NewDB()

	if err := common.MigrateDatabaseUp(ctx, moduleName, pool, embedFS, migrationsDir); err != nil {
		panic(err)
	}
}

// NewDB returns a *pgxpool.Pool shared for the lifetime of the test process, creating it once via
// sync.Once. Every caller reuses this same pool instead of each opening its own. A background
// goroutine closes the pool once CloseDB is called.
func NewDB() *pgxpool.Pool {
	dbPoolOnce.Do(func() {
		config, err := pgxpool.ParseConfig(common.NewConfig().DB.URL)
		if err != nil {
			panic(err)
		}

		pool, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			panic(err)
		}

		dbPool = pool
	})

	return dbPool
}

// CloseDB signals the background goroutine started by NewDB to close the shared pool. Safe to call
// even if NewDB was never invoked, and safe to call more than once.
func CloseDB() {
	dbPool.Close()
}
