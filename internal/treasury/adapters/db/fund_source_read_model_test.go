//go:build integration

package db_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	repoDb "sumni-finance-backend/internal/treasury/adapters/db"
	"sumni-finance-backend/internal/treasury/app/query"
	"sumni-finance-backend/internal/treasury/domain"
)

func TestListFundSources_IncludesBankFundSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, cleanup := testutils.NewDB(ctx)
	defer cleanup()
	repo := repoDb.NewFundSourceRepository(db)
	readModel := repoDb.NewFundSourceReadModel(db)

	fs := mustCreateBankFundSource(t, ctx, repo, decimal.NewFromInt(1_000_000))

	views, err := readModel.ListFundSources(ctx)
	require.NoError(t, err)

	got := findFundSourceViewByUUID(views, fs.UUID())
	require.NotNil(t, got, "expected bank fund source %s to appear in list", fs.UUID())

	want := expectedFundSourceView(fs)
	if diff := cmp.Diff(want, *got, cmpOpts...); diff != "" {
		t.Errorf("bank fund source view mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, got.CashMetadata)
}

func TestListFundSources_IncludesCashFundSource(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, cleanup := testutils.NewDB(ctx)
	defer cleanup()
	repo := repoDb.NewFundSourceRepository(db)
	readModel := repoDb.NewFundSourceReadModel(db)

	fs := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(500_000))

	views, err := readModel.ListFundSources(ctx)
	require.NoError(t, err)

	got := findFundSourceViewByUUID(views, fs.UUID())
	require.NotNil(t, got, "expected cash fund source %s to appear in list", fs.UUID())

	want := expectedFundSourceView(fs)
	if diff := cmp.Diff(want, *got, cmpOpts...); diff != "" {
		t.Errorf("cash fund source view mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, got.BankMetadata)
}

func TestListFundSources_ReturnsAllSavedFundSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, cleanup := testutils.NewDB(ctx)
	defer cleanup()
	repo := repoDb.NewFundSourceRepository(db)
	readModel := repoDb.NewFundSourceReadModel(db)

	cash := mustCreateCashFundSource(t, ctx, repo, decimal.NewFromInt(100_000))
	bank := mustCreateBankFundSource(t, ctx, repo, decimal.NewFromInt(200_000))

	views, err := readModel.ListFundSources(ctx)
	require.NoError(t, err)

	assert.NotNil(t, findFundSourceViewByUUID(views, cash.UUID()), "cash fund source should appear in list")
	assert.NotNil(t, findFundSourceViewByUUID(views, bank.UUID()), "bank fund source should appear in list")
}

func mustCreateBankFundSource(t *testing.T, ctx context.Context, repo *repoDb.FundSourceRepo, balanceAmt decimal.Decimal) *domain.FundSource {
	t.Helper()

	balance, err := shared.NewMoney(balanceAmt, vnd)
	require.NoError(t, err)

	metadata, err := fundSourceFactory.NewBankMetadata(ctx, "VCB", gofakeit.Numerify("##########"), gofakeit.Name())
	require.NoError(t, err)

	fs, err := fundSourceFactory.NewFundSource("Bank Account", domain.FundSourceTypeBank, balance, vnd, metadata)
	require.NoError(t, err)

	require.NoError(t, repo.SaveFundSource(ctx, fs))
	return fs
}

func findFundSourceViewByUUID(views []query.FundSourceView, uuid domain.FundSourceUUID) *query.FundSourceView {
	for i := range views {
		if views[i].UUID == uuid {
			return &views[i]
		}
	}
	return nil
}

func expectedFundSourceView(fs *domain.FundSource) query.FundSourceView {
	view := query.FundSourceView{
		UUID:             fs.UUID(),
		Name:             fs.Name(),
		SourceType:       fs.SourceType(),
		Balance:          fs.Balance().Amount(),
		AvailableBalance: fs.AvailableBalance().Amount(),
		Currency:         fs.Currency(),
		CreatedAt:        fs.Audit().CreatedAt(),
		CreatedBy:        fs.Audit().CreatedBy(),
		UpdatedAt:        fs.Audit().UpdatedAt(),
		UpdatedBy:        fs.Audit().UpdatedBy(),
	}

	if bankMeta, ok := fs.BankMetadata(); ok {
		view.BankMetadata = &query.FundSourceBankMetadataView{
			BankName:      bankMeta.BankInfo().Name(),
			BankCode:      bankMeta.BankInfo().BankCode(),
			BankShortName: bankMeta.BankInfo().ShortName(),
			BankBin:       bankMeta.BankInfo().Bin(),
			BankLogoUrl:   bankMeta.BankInfo().LogoUrl(),
			BankIconUrl:   bankMeta.BankInfo().IconUrl(),
			AccountNumber: bankMeta.AccountNumber(),
			AccountOwner:  bankMeta.AccountOwner(),
		}
	}

	if cashMeta, ok := fs.CashMetadata(); ok {
		view.CashMetadata = &query.FundSourceCashMetadataView{
			OwnerName: cashMeta.OwnerName(),
		}
	}

	return view
}
