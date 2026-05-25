//go:build integration

package db_test

import (
	"context"
	"testing"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/finance/adapters/db"
	"sumni-finance-backend/internal/finance/app/models"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestOfficeRepo_SaveOffice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database := testutils.NewDB(t)
	officeRepo := db.NewOfficeRepo(database)

	t.Run("insert new office", func(t *testing.T) {
		office := models.Office{
			UUID: models.OfficeUUID{UUID: common.NewUUIDv7()},
			Name: testutils.RandomString(10),
		}
		err := officeRepo.SaveOffice(ctx, office)
		require.NoError(t, err)

		got, err := officeRepo.OfficeByUUID(ctx, office.UUID)
		require.NoError(t, err)
		if diff := cmp.Diff(office, got); diff != "" {
			t.Errorf("office mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("update existing office name", func(t *testing.T) {
		original := models.Office{
			UUID: models.OfficeUUID{UUID: common.NewUUIDv7()},
			Name: testutils.RandomString(10),
		}
		err := officeRepo.SaveOffice(ctx, original)
		require.NoError(t, err)

		updated := models.Office{
			UUID: original.UUID,
			Name: testutils.RandomString(10),
		}
		err = officeRepo.SaveOffice(ctx, updated)
		require.NoError(t, err)

		got, err := officeRepo.OfficeByUUID(ctx, original.UUID)
		require.NoError(t, err)
		if diff := cmp.Diff(updated, got); diff != "" {
			t.Errorf("office mismatch (-want +got):\n%s", diff)
		}
	})
}
