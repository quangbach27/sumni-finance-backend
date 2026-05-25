//go:build integration

package db_test

import (
	"embed"
	"os"
	"testing"

	"sumni-finance-backend/internal/common/testutils"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func TestMain(m *testing.M) {
	testutils.RunMigrations("finances", embedMigrations, "migrations")
	os.Exit(m.Run())
}
