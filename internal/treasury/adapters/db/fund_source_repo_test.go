package db_test

import (
	"context"
	"testing"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/treasury/adapters/db"
	"sumni-finance-backend/internal/treasury/adapters/db/dbmodels"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var vnd = shared.MustNewCurrency("VND")

func TestSaveFundSource(t *testing.T) {
	t.Parallel()

	t.Run("bank_fund_source", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		pgxDb := testutils.NewDB()
		repo := db.NewFundSourceRepository(pgxDb)

		balance, err := shared.NewMoney(decimal.NewFromInt(500_000), vnd)
		require.NoError(t, err)

		metadata, err := domain.NewBankMetadata(
			gofakeit.Numerify("##########"),
			gofakeit.Name(),
			domain.BankInfoData{
				Name:      "Vietcombank",
				Bin:       "970436",
				ShortName: "VCB",
			},
		)
		require.NoError(t, err)

		fs, err := domain.NewFundSource("Techcombank-SRB", domain.FundSourceTypeBank, balance, vnd, metadata)
		require.NoError(t, err)

		tenantContext, err := shared.NewTenantContext("tenant-1", "office-1")
		require.NoError(t, err)
		require.NoError(t, repo.CreateFundSource(ctx, tenantContext, fs))

		queries := dbmodels.New(pgxDb)
		actualRow, err := queries.GetFundSourceByUUID(ctx, fs.UUID())
		require.NoError(t, err)

		if diff := cmp.Diff(
			fundSourceDomainToExpectedRow(fs),
			actualRow,
			cmp.AllowUnexported(
				domain.BankInfo{},
			),
		); diff != "" {
			t.Errorf("saved row mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("cash_fund_source", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		pgxDb := testutils.NewDB()
		repo := db.NewFundSourceRepository(pgxDb)

		balance, err := shared.NewMoney(decimal.NewFromInt(500_000), vnd)
		require.NoError(t, err)

		metadata, err := domain.NewCashMetadata(gofakeit.Name())
		require.NoError(t, err)

		fs, err := domain.NewFundSource("Cash-Wallet", domain.FundSourceTypeCash, balance, vnd, metadata)
		require.NoError(t, err)

		tenantContext, err := shared.NewTenantContext("tenant-1", "office-1")
		require.NoError(t, err)
		require.NoError(t, repo.CreateFundSource(ctx, tenantContext, fs))

		queries := dbmodels.New(pgxDb)
		actualRow, err := queries.GetFundSourceByUUID(ctx, fs.UUID())
		require.NoError(t, err)

		if diff := cmp.Diff(
			fundSourceDomainToExpectedRow(fs),
			actualRow,
			cmp.AllowUnexported(
				domain.BankInfo{},
			),
		); diff != "" {
			t.Errorf("saved row mismatch (-want +got):\n%s", diff)
		}
	})
}

func fundSourceDomainToExpectedRow(fs *domain.FundSource) dbmodels.TreasuryFundSource {
	row := dbmodels.TreasuryFundSource{
		FundSourceUuid:   fs.UUID(),
		Name:             fs.Name(),
		SourceType:       fs.SourceType(),
		Balance:          fs.Balance().Amount(),
		AvailableBalance: fs.AvailableBalance().Amount(),
		Currency:         fs.Currency(),
		TenantID:         "tenant-1",
		OfficeID:         "office-1",
	}

	if bankMeta, ok := fs.BankMetadata(); ok {
		row.BankInfo = bankMeta.BankInfo()
		row.Bin = common.ToPtr(bankMeta.BankInfo().Bin())
		row.BankAccountNumber = common.ToPtr(bankMeta.AccountNumber())
		row.BankAccountOwner = common.ToPtr(bankMeta.AccountOwner())
	}

	if cashMeta, ok := fs.CashMetadata(); ok {
		row.CashOwner = common.ToPtr(cashMeta.OwnerName())
	}

	return row
}
