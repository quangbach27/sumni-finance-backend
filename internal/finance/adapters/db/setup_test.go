//go:build integration

package db_test

import (
	"embed"
	"os"
	"testing"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

var vnd = shared.MustNewCurrency("VND")

func TestMain(m *testing.M) {
	testutils.RunMigrations("finances", embedMigrations, "migrations")
	os.Exit(m.Run())
}
