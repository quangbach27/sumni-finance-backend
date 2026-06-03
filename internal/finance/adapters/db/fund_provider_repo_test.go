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
	t.Parallel()
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

	fp := validFundProvider(t, office.UUID, decimal.NewFromInt(10_000_000), domain.FundProviderTypeBank)

	err = fpRepo.SaveFundProvider(ctx, fp)
	require.NoError(t, err)

	queries := dbmodels.New(database)

	fpModel, err := queries.FundProviderByUUID(ctx, fp.UUID())
	require.NoError(t, err)

	bankMetadata, ok := fp.BankMetadata()
	require.True(t, ok)

	if diff := cmp.Diff(
		dbmodels.FinancesFundProvider{
			FundProviderUuid:  fp.UUID(),
			Name:              fp.Name(),
			Balance:           fp.Balance().Amount(),
			AvailableBalance:  fp.AvailableBalance().Amount(),
			Currency:          fp.Currency(),
			FundProviderType:  fp.Type(),
			BankInfo:          bankMetadata.BankInfo(),
			BankAccountNumber: common.ToPtr(bankMetadata.AccountNumber()),
			BankAccountOwner:  common.ToPtr(bankMetadata.AccountOwner()),
			OfficeUuid:        fp.OfficeUUID(),
		},
		fpModel,
		cmpopts.EquateComparable(shared.SharedTypes...),
	); diff != "" {
		t.Errorf("fund provider mismatch (-want +got):\n%s", diff)
	}
}

func TestInsertFundProvider_Cash(t *testing.T) {
	t.Parallel()
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

	fp := validFundProvider(t, office.UUID, decimal.NewFromInt(10_000_000), domain.FundProviderTypeCash)
	cashMetadata, ok := fp.CashMetadata()
	require.True(t, ok)

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
			AvailableBalance: fp.AvailableBalance().Amount(),
			Currency:         fp.Currency(),
			FundProviderType: fp.Type(),
			OfficeUuid:       fp.OfficeUUID(),
			CashOwnerName:    common.ToPtr(cashMetadata.OwnerName()),
		},
		fpModel,
		cmpopts.EquateComparable(shared.SharedTypes...),
	); diff != "" {
		t.Errorf("fund provider mismatch (-want +got):\n%s", diff)
	}
}

func validFundProvider(
	t *testing.T,
	officeUUID models.OfficeUUID,
	initBalanceAmount decimal.Decimal,
	fpType domain.FundProviderType,
) *domain.FundProvider {
	t.Helper()

	vnd := shared.MustNewCurrency("VND")

	fp, err := domain.NewFundProvider(
		testutils.RandomString(10),
		officeUUID,
		fpType,
		shared.UnmarshalMoney(initBalanceAmount, vnd),
		validFundProviderMetadata(t, fpType),
	)
	require.NoError(t, err)

	return fp
}

func validFundProviderMetadata(t *testing.T, fpType domain.FundProviderType) domain.FundProviderMetadata {
	t.Helper()

	var metadata domain.FundProviderMetadata
	var err error
	switch fpType {
	case domain.FundProviderTypeBank:
		metadata, err = domain.NewFundProviderBankMetadata(
			testutils.RandomBankInfo(t),
			testutils.RandomString(10),
			testutils.RandomString(10),
		)
		require.NoError(t, err)
	case domain.FundProviderTypeCash:
		metadata, err = domain.NewCashFundProviderMetadata(testutils.RandomString(10))
		require.NoError(t, err)
	default:
		t.Fatalf("unkown fund provider type: %s", fpType.String())
	}

	return metadata
}
