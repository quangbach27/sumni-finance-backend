//go:build integration

package db

import (
	"embed"
	"os"
	"testing"

	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/treasury"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func TestMain(m *testing.M) {
	module := treasury.Module{}

	testutils.RunMigrations(string(module.Name()), embedMigrations, "migrations")
	os.Exit(m.Run())
}
