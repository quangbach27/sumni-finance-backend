//go:build integration

package db

import (
	"embed"
	"os"
	"testing"

	"sumni-finance-backend/internal/common/testutils"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func TestMain(m *testing.M) {
	testutils.RunMigrations("treasury", embedMigrations, "migrations")
	os.Exit(m.Run())
}
