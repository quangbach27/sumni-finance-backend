//go:build integration

package db_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/finance/adapters/db"
	"sumni-finance-backend/internal/finance/adapters/db/dbmodels"
	"sumni-finance-backend/internal/finance/app/models"
	"sumni-finance-backend/internal/finance/domain"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertWallet_WithAllocation(t *testing.T) {
	t.Parallel()

	database := testutils.NewDB(t)
	ctx := context.Background()

	walletRepo := db.NewWalletRepository(database)
	officeRepo := db.NewOfficeRepo(database)
	fpRepo := db.NewFundProviderRepository(database)

	office := validOffice(t)

	err := officeRepo.SaveOffice(ctx, office)
	require.NoError(t, err)

	fp1 := validFundProvider(t, office.UUID, decimal.NewFromInt(1_000_000), domain.FundProviderTypeBank)
	fp2 := validFundProvider(t, office.UUID, decimal.NewFromInt(1_000_000), domain.FundProviderTypeCash)

	err = fpRepo.SaveFundProvider(ctx, fp1)
	require.NoError(t, err)
	err = fpRepo.SaveFundProvider(ctx, fp2)
	require.NoError(t, err)

	allocationAmount := shared.UnmarshalMoney(decimal.NewFromInt(500_000), vnd)

	var wallet *domain.Wallet

	err = walletRepo.SaveWallet(
		ctx,
		[]domain.FundProviderUUID{fp1.UUID(), fp2.UUID()},
		func(fundProviders map[domain.FundProviderUUID]*domain.FundProvider) (*domain.Wallet, error) {
			wallet = validWallet(
				t,
				office.UUID,
				domain.NewFundProviderAllocationData{
					FundProvider:     fundProviders[fp1.UUID()],
					AllocationAmount: allocationAmount,
				},
				domain.NewFundProviderAllocationData{
					FundProvider:     fundProviders[fp2.UUID()],
					AllocationAmount: allocationAmount,
				},
			)
			return wallet, nil
		},
	)
	require.NoError(t, err)

	gotWallet, err := walletRepo.GetWalletWithAllocations(ctx, wallet.UUID())
	require.NoError(t, err)

	assertWallet(t, wallet, gotWallet)
}

func TestInsertWallet_WithoutAllocation(t *testing.T) {
	t.Parallel()

	database := testutils.NewDB(t)
	ctx := context.Background()

	walletRepo := db.NewWalletRepository(database)
	officeRepo := db.NewOfficeRepo(database)

	office := validOffice(t)

	err := officeRepo.SaveOffice(ctx, office)
	require.NoError(t, err)

	var wallet *domain.Wallet

	err = walletRepo.SaveWallet(
		ctx,
		nil,
		func(fundProviders map[domain.FundProviderUUID]*domain.FundProvider) (*domain.Wallet, error) {
			wallet = validWallet(
				t,
				office.UUID,
			)
			return wallet, nil
		},
	)
	require.NoError(t, err)

	gotWallet, err := walletRepo.GetWalletWithAllocations(ctx, wallet.UUID())
	require.NoError(t, err)

	assertWallet(t, wallet, gotWallet)
}

type walletSaveResult struct {
	wallet *domain.Wallet
	err    error
}

func TestInsertWallet_ConcurrentWalletAllocateSameFundProvider(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	database := testutils.NewDB(t)
	walletRepo := db.NewWalletRepository(database)
	fpRepo := db.NewFundProviderRepository(database)
	queries := dbmodels.New(database)

	office := assertInsertOffice(t, database)

	const (
		allocationAmount    = 200_000
		fundProviderBalance = 1_000_000
		total               = fundProviderBalance / allocationAmount // 2
	)

	fp := validFundProvider(t, office.UUID, decimal.NewFromInt(fundProviderBalance), domain.FundProviderTypeBank)
	err := fpRepo.SaveFundProvider(ctx, fp)
	require.NoError(t, err)

	allocationMoney := shared.UnmarshalMoney(decimal.NewFromInt(allocationAmount), vnd)
	results := make(chan walletSaveResult, total)

	for i := 0; i < total; i++ {
		go func() {
			var w *domain.Wallet
			saveErr := walletRepo.SaveWallet(
				ctx,
				[]domain.FundProviderUUID{fp.UUID()},
				func(fundProviders map[domain.FundProviderUUID]*domain.FundProvider) (*domain.Wallet, error) {
					var internalErr error
					w, internalErr = newRandomWallet(office.UUID, domain.NewFundProviderAllocationData{
						FundProvider:     fundProviders[fp.UUID()],
						AllocationAmount: allocationMoney,
					})

					return w, internalErr
				})

			results <- walletSaveResult{
				wallet: w,
				err:    saveErr,
			}
		}()
	}

	deadline := time.After(10 * time.Second)
	for i := 0; i < total; i++ {
		select {
		case r := <-results:
			assert.NoError(t, r.err)
			assert.NotNil(t, r.wallet)
			if r.wallet != nil {
				allocationsDb, err := queries.FundProvidersWithAllocations(ctx, r.wallet.UUID())
				assert.NoError(t, err)

				wantAvailableBalance := decimal.NewFromInt(int64(fundProviderBalance - allocationAmount*(i+1)))
				wantAllocationAmount := decimal.NewFromInt(int64(allocationAmount))

				assert.True(t, allocationsDb[0].FinancesFundProvider.AvailableBalance.Equal(wantAvailableBalance))
				assert.True(t, allocationsDb[0].FinancesFundProviderAllocation.AllocationAmount.Equal(wantAllocationAmount))
			}
		case <-deadline:
			t.Fatalf("timed out after collecting %d/%d results", i, total)
		}
	}

	newFp, err := queries.FundProviderByUUID(ctx, fp.UUID())
	require.NoError(t, err)

	assert.Equal(t, newFp.AvailableBalance, decimal.NewFromInt(0))
}

func validWallet(t *testing.T, officeUUID models.OfficeUUID, allocationsData ...domain.NewFundProviderAllocationData) *domain.Wallet {
	t.Helper()

	wallet, err := domain.NewWallet(
		testutils.RandomString(12),
		testutils.RandomString(12),
		vnd,
		officeUUID,
		allocationsData,
	)
	require.NoError(t, err)

	return wallet
}

func newRandomWallet(officeUUID models.OfficeUUID, allocationsData ...domain.NewFundProviderAllocationData) (*domain.Wallet, error) {
	return domain.NewWallet(
		testutils.RandomString(12),
		testutils.RandomString(12),
		vnd,
		officeUUID,
		allocationsData,
	)
}

func assertWallet(t *testing.T, want, got *domain.Wallet) {
	t.Helper()

	require.NotNil(t, want)
	require.NotNil(t, got)

	assert.Equal(t, want.UUID(), got.UUID())
	assert.Equal(t, want.Name(), got.Name())
	assert.Equal(t, want.Description(), got.Description())
	assert.True(t, want.Balance().Equal(got.Balance()), "balance mismatch: want %s, got %s", want.Balance(), got.Balance())
	assert.Equal(t, want.Currency(), got.Currency())
	assert.Equal(t, want.OfficeUUID(), got.OfficeUUID())

	wantAllocations := want.FundProviderAllocations()
	gotAllocations := got.FundProviderAllocations()

	require.Len(t, gotAllocations, len(wantAllocations))

	if len(gotAllocations) != 0 {
		sortAllocations := func(allocs []*domain.FundProviderAllocation) {
			sort.Slice(allocs, func(i, j int) bool {
				return allocs[i].FundProvider().UUID().String() < allocs[j].FundProvider().UUID().String()
			})
		}
		sortAllocations(wantAllocations)
		sortAllocations(gotAllocations)

		for i := range wantAllocations {
			w, g := wantAllocations[i], gotAllocations[i]
			assert.Equal(t, w.FundProvider().UUID(), g.FundProvider().UUID())
			assert.True(t, w.Amount().Equal(g.Amount()), "allocation[%d] amount mismatch: want %s, got %s", i, w.Amount(), g.Amount())
		}
	}
}
