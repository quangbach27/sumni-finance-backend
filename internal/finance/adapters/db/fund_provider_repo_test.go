//go:build integration

package db_test

import (
	"context"
	"testing"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/finance/adapters/db"
	"sumni-finance-backend/internal/finance/adapters/db/dbmodels"
	"sumni-finance-backend/internal/finance/app/models"
	"sumni-finance-backend/internal/finance/domain"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/go-cmp/cmp"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestInsertFundProvider_Bank(t *testing.T) {
	ctx := context.Background()

	database := testutils.NewDB(t)

	fpRepo := db.NewFundProviderRepository(database)
	officeRepo := db.NewOfficeRepo(database)

	office := models.Office{
		UUID: models.OfficeUUID{UUID: common.NewUUIDv7()},
		Name: testutils.RandomString(10),
	}

	err := officeRepo.SaveOffice(ctx, office)
	require.NoError(t, err)

	bankMetadata, err := domain.NewBankFundProviderMetadata(
		testutils.RandomString(10),
		testutils.RandomString(10),
		testutils.RandomString(10),
	)
	require.NoError(t, err)

	vnd := shared.MustNewCurrency("VND")

	initBalance, err := shared.NewMoney(decimal.NewFromInt(10_000_000), vnd)

	fp, err := domain.NewFundProvider(
		testutils.RandomString(10),
		office.UUID,
		domain.FundProviderTypeBank,
		initBalance,
		bankMetadata,
	)

	require.NoError(t, err)

	err = fpRepo.SaveFundProvider(ctx, fp)
	require.NoError(t, err)

	queries := dbmodels.New(database)

	fpModel, err := queries.FundProviderByUUID(ctx, fp.UUID())
	require.NoError(t, err)

	if diff := cmp.Diff(
		dbmodels.FinancesFundProvider{
			FundProviderUuid:  fp.UUID(),
			Name:              fp.Name(),
			Balance:           fp.Balance().Amount(),
			AvailableBalance:  fp.AvailableMoney().Amount(),
			Currency:          fp.Currency(),
			FundProviderType:  fp.Type(),
			BankName:          common.Ptr(bankMetadata.BankName()),
			BankAccountNumber: common.Ptr(bankMetadata.AccountNumber()),
			BankAccountOwner:  common.Ptr(bankMetadata.AccountOwner()),
			OfficeUuid:        fp.OfficeUUID(),
		},
		fpModel,
		cmpopts.EquateComparable(shared.SharedTypes...),
	); diff != "" {
		t.Errorf("fund provider mismatch (-want +got):\n%s", diff)
	}
}

func TestInsertFundProvider_Cash(t *testing.T) {
	ctx := context.Background()

	database := testutils.NewDB(t)

	fpRepo := db.NewFundProviderRepository(database)
	officeRepo := db.NewOfficeRepo(database)

	office := models.Office{
		UUID: models.OfficeUUID{UUID: common.NewUUIDv7()},
		Name: testutils.RandomString(10),
	}

	err := officeRepo.SaveOffice(ctx, office)
	require.NoError(t, err)

	cashMetadata, err := domain.NewCashFundProviderMetadata(testutils.RandomString(10))
	require.NoError(t, err)

	vnd := shared.MustNewCurrency("VND")

	initBalance, err := shared.NewMoney(decimal.NewFromInt(10_000_000), vnd)

	fp, err := domain.NewFundProvider(
		testutils.RandomString(10),
		office.UUID,
		domain.FundProviderTypeCash,
		initBalance,
		cashMetadata,
	)
	require.NoError(t, err)

	err = fpRepo.SaveFundProvider(ctx, fp)
	require.NoError(t, err)

	queries := dbmodels.New(database)

	fpModel, err := queries.FundProviderByUUID(ctx, fp.UUID())
	require.NoError(t, err)

	if diff := cmp.Diff(
		dbmodels.FinancesFundProvider{
			FundProviderUuid: fp.UUID(),
			Name:             fp.Name(),
			Balance:          fp.Balance().Amount(),
			AvailableBalance: fp.AvailableMoney().Amount(),
			Currency:         fp.Currency(),
			FundProviderType: fp.Type(),
			OfficeUuid:       fp.OfficeUUID(),
			CashOwnerName:    common.Ptr(cashMetadata.OwnerName()),
		},
		fpModel,
	); diff != "" {
		t.Errorf("fund provider mismatch (-want +got):\n%s", diff)
	}
}
