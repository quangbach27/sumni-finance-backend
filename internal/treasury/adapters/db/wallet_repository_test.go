//go:build integration

package db_test

import (
	"context"
	"sync"
	"testing"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/common/testutils"
	"sumni-finance-backend/internal/treasury/adapters/db"
	"sumni-finance-backend/internal/treasury/adapters/db/dbmodels"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCreateWallet_SavesWalletAndAllocations(t *testing.T) {
	ctx := context.Background()
	pgxDB := testutils.NewDB()
	fundSourceRepo := db.NewFundSourceRepository(pgxDB)
	walletRepo := db.NewWalletRepository(pgxDB)

	tenantContext, err := shared.NewTenantContext("tenant-1", "office-1")
	require.NoError(t, err)

	balance, err := shared.NewMoney(decimal.NewFromInt(1_000_000), vnd)
	require.NoError(t, err)

	metadata, err := domain.NewCashMetadata("Finance Ops")
	require.NoError(t, err)

	fundSource, err := domain.NewFundSource("Cash Reserve", domain.FundSourceTypeCash, balance, vnd, metadata)
	require.NoError(t, err)

	require.NoError(t, fundSourceRepo.CreateFundSource(ctx, tenantContext, fundSource))

	wallet, err := domain.NewWallet("Operations Wallet", vnd)
	require.NoError(t, err)

	allocationAmount, err := shared.NewMoney(decimal.NewFromInt(250_000), vnd)
	require.NoError(t, err)

	require.NoError(t, walletRepo.CreateWallet(
		ctx,
		tenantContext,
		[]domain.FundSourceUUID{fundSource.UUID()},
		func(fundSources map[domain.FundSourceUUID]*domain.FundSource) (*domain.Wallet, error) {
			if err := wallet.Allocate(fundSources[fundSource.UUID()], allocationAmount); err != nil {
				return nil, err
			}
			return wallet, nil
		},
	))

	queries := dbmodels.New(pgxDB)
	actualWallet, err := queries.GetWalletByUUID(ctx, dbmodels.GetWalletByUUIDParams{
		WalletUuid: wallet.UUID(),
		TenantID:   tenantContext.TenantID(),
		OfficeID:   tenantContext.OfficeID(),
	})
	require.NoError(t, err)

	if diff := cmp.Diff(walletDomainToExpectedRow(wallet), actualWallet); diff != "" {
		t.Fatalf("saved wallet mismatch (-want +got):\n%s", diff)
	}

	allocations, err := queries.GetAllocationsByWalletUUID(ctx, wallet.UUID())
	require.NoError(t, err)

	expectedWalletAllocations := []dbmodels.TreasuryWalletAllocation{{
		WalletUuid:     wallet.UUID(),
		FundSourceUuid: fundSource.UUID(),
		Balance:        allocationAmount.Amount(),
	}}

	actualWalletAllocations := make([]dbmodels.TreasuryWalletAllocation, 0, len(allocations))
	for _, allocation := range allocations {
		actualWalletAllocations = append(actualWalletAllocations, allocation.TreasuryWalletAllocation)
	}

	if diff := cmp.Diff(expectedWalletAllocations, actualWalletAllocations); diff != "" {
		t.Fatalf("saved wallet allocations mismatch (-want +got):\n%s", diff)
	}
}

func TestCreateWallet_Concurrent(t *testing.T) {
	ctx := context.Background()
	pgxDB := testutils.NewDB()
	fundSourceRepo := db.NewFundSourceRepository(pgxDB)
	walletRepo := db.NewWalletRepository(pgxDB)

	tenantContext, err := shared.NewTenantContext("tenant-1", "office-1")
	require.NoError(t, err)

	initialBalance, err := shared.NewMoney(decimal.NewFromInt(1_000_000), vnd)
	require.NoError(t, err)

	metadata, err := domain.NewCashMetadata("Finance Ops")
	require.NoError(t, err)

	fundSource, err := domain.NewFundSource("Cash Reserve", domain.FundSourceTypeCash, initialBalance, vnd, metadata)
	require.NoError(t, err)

	require.NoError(t, fundSourceRepo.CreateFundSource(ctx, tenantContext, fundSource))

	const goroutines = 8
	const reserveAmount = 250_000
	var wg sync.WaitGroup
	resultsCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			allocationAmount, err := shared.NewMoney(decimal.NewFromInt(reserveAmount), vnd)
			if err != nil {
				resultsCh <- err
				return
			}

			if err = walletRepo.CreateWallet(ctx, tenantContext, []domain.FundSourceUUID{fundSource.UUID()}, func(fundSources map[domain.FundSourceUUID]*domain.FundSource) (*domain.Wallet, error) {
				wallet, err := domain.NewWallet("Operations Wallet "+string(rune('A'+i)), vnd)
				if err != nil {
					return nil, err
				}
				if err = wallet.Allocate(fundSources[fundSource.UUID()], allocationAmount); err != nil {
					return nil, err
				}
				return wallet, nil
			}); err != nil {
				resultsCh <- err
				return
			}

			resultsCh <- nil
		}(i)
	}

	wg.Wait()
	close(resultsCh)

	successCount := 0
	failureCount := 0
	for err := range resultsCh {
		if err != nil {
			failureCount++
			continue
		}
		successCount++
	}

	require.Equal(t, 4, successCount, "expected exactly four goroutines to reserve successfully")
	require.Equal(t, 4, failureCount, "expected exactly four goroutines to fail once the balance is exhausted")

	actualFundSource, err := getFundSourceByUUID(ctx, pgxDB, fundSource.UUID())
	require.NoError(t, err)

	expectedAvailableBalance := decimal.NewFromInt(int64(1_000_000 - (reserveAmount * successCount)))
	require.True(t, actualFundSource.AvailableBalance.Equal(expectedAvailableBalance), "expected available balance %s, got %s", expectedAvailableBalance, actualFundSource.AvailableBalance)
}

func TestLinkFundSources_SavesAllocationsAndUpdatesBalances(t *testing.T) {
	ctx := context.Background()
	pgxDB := testutils.NewDB()
	fundSourceRepo := db.NewFundSourceRepository(pgxDB)
	walletRepo := db.NewWalletRepository(pgxDB)

	tenantContext, err := shared.NewTenantContext("tenant-1", "office-1")
	require.NoError(t, err)

	initialBalance, err := shared.NewMoney(decimal.NewFromInt(500_000), vnd)
	require.NoError(t, err)

	metadata, err := domain.NewCashMetadata("Finance Ops")
	require.NoError(t, err)

	fundSource, err := domain.NewFundSource("Cash Reserve", domain.FundSourceTypeCash, initialBalance, vnd, metadata)
	require.NoError(t, err)

	require.NoError(t, fundSourceRepo.CreateFundSource(ctx, tenantContext, fundSource))

	wallet, err := domain.NewWallet("Ops Wallet", vnd)
	require.NoError(t, err)

	require.NoError(t, walletRepo.CreateWallet(
		ctx,
		tenantContext,
		nil,
		func(_ map[domain.FundSourceUUID]*domain.FundSource) (*domain.Wallet, error) {
			return wallet, nil
		},
	))

	allocationAmount, err := shared.NewMoney(decimal.NewFromInt(200_000), vnd)
	require.NoError(t, err)

	err = walletRepo.LinkFundSources(ctx, tenantContext, wallet.UUID(), []domain.FundSourceAllocationData{{
		FundSourceUUID:  fundSource.UUID(),
		AllocatedAmount: allocationAmount.Amount(),
	}})
	require.NoError(t, err)

	gotWallet, err := walletRepo.GetWallet(ctx, tenantContext, wallet.UUID())
	require.NoError(t, err)
	require.True(t, gotWallet.Balance().Amount().Equal(decimal.NewFromInt(200_000)))

	allocations, err := getWalletAllocationsByWalletUUID(ctx, pgxDB, wallet.UUID())
	require.NoError(t, err)
	require.Len(t, allocations, 1)
	require.Equal(t, fundSource.UUID(), allocations[0].FundSourceUuid)
	require.True(t, allocations[0].Balance.Equal(allocationAmount.Amount()))

	updatedFundSource, err := getFundSourceByUUID(ctx, pgxDB, fundSource.UUID())
	require.NoError(t, err)
	require.True(t, updatedFundSource.AvailableBalance.Equal(decimal.NewFromInt(300_000)))
}

func TestGetWallet_ReturnsWalletWithoutAllocations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pgxDB := testutils.NewDB()
	walletRepo := db.NewWalletRepository(pgxDB)

	tenantContext, err := shared.NewTenantContext("tenant-1", "office-1")
	require.NoError(t, err)

	wallet, err := domain.NewWallet("Savings Wallet", vnd)
	require.NoError(t, err)

	require.NoError(t, walletRepo.CreateWallet(
		ctx,
		tenantContext,
		nil,
		func(_ map[domain.FundSourceUUID]*domain.FundSource) (*domain.Wallet, error) {
			return wallet, nil
		},
	))

	got, err := walletRepo.GetWallet(ctx, tenantContext, wallet.UUID())
	require.NoError(t, err)
	require.Equal(t, wallet.UUID(), got.UUID())
	require.Equal(t, wallet.Name(), got.Name())
	require.True(t, wallet.Balance().Amount().Equal(got.Balance().Amount()))
	require.Equal(t, wallet.Balance().Currency(), got.Balance().Currency())
	require.Empty(t, got.Allocations())
}

func TestGetWallet_ReturnsWalletWithAllocations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pgxDB := testutils.NewDB()
	fundSourceRepo := db.NewFundSourceRepository(pgxDB)
	walletRepo := db.NewWalletRepository(pgxDB)

	tenantContext, err := shared.NewTenantContext("tenant-1", "office-1")
	require.NoError(t, err)

	initialBalance, err := shared.NewMoney(decimal.NewFromInt(500_000), vnd)
	require.NoError(t, err)

	metadata, err := domain.NewCashMetadata("Finance Ops")
	require.NoError(t, err)

	fundSource, err := domain.NewFundSource("Cash Reserve", domain.FundSourceTypeCash, initialBalance, vnd, metadata)
	require.NoError(t, err)

	require.NoError(t, fundSourceRepo.CreateFundSource(ctx, tenantContext, fundSource))

	allocationAmount, err := shared.NewMoney(decimal.NewFromInt(200_000), vnd)
	require.NoError(t, err)

	wallet, err := domain.NewWallet("Ops Wallet", vnd)
	require.NoError(t, err)

	require.NoError(t, walletRepo.CreateWallet(
		ctx,
		tenantContext,
		[]domain.FundSourceUUID{fundSource.UUID()},
		func(fundSources map[domain.FundSourceUUID]*domain.FundSource) (*domain.Wallet, error) {
			if err := wallet.Allocate(fundSources[fundSource.UUID()], allocationAmount); err != nil {
				return nil, err
			}
			return wallet, nil
		},
	))

	got, err := walletRepo.GetWallet(ctx, tenantContext, wallet.UUID())
	require.NoError(t, err)
	require.Equal(t, wallet.UUID(), got.UUID())
	require.Equal(t, wallet.Name(), got.Name())
	require.True(t, wallet.Balance().Amount().Equal(got.Balance().Amount()))
	require.Equal(t, wallet.Balance().Currency(), got.Balance().Currency())
}

func getWalletAllocationsByWalletUUID(ctx context.Context, pgxDB *pgxpool.Pool, walletUUID domain.WalletUUID) ([]dbmodels.TreasuryWalletAllocation, error) {
	rows, err := pgxDB.Query(ctx, "SELECT wallet_uuid, fund_source_uuid, balance FROM treasury.wallet_allocations WHERE wallet_uuid = $1 ORDER BY fund_source_uuid", walletUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocations []dbmodels.TreasuryWalletAllocation
	for rows.Next() {
		var allocation dbmodels.TreasuryWalletAllocation
		if err := rows.Scan(&allocation.WalletUuid, &allocation.FundSourceUuid, &allocation.Balance); err != nil {
			return nil, err
		}
		allocations = append(allocations, allocation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return allocations, nil
}

func getFundSourceByUUID(ctx context.Context, pgxDB *pgxpool.Pool, fundSourceUUID domain.FundSourceUUID) (dbmodels.TreasuryFundSource, error) {
	row := pgxDB.QueryRow(ctx, "SELECT fund_source_uuid, name, source_type, currency, balance, available_balance, bank_info, bin, bank_account_number, bank_account_owner, cash_owner, tenant_id, office_id FROM treasury.fund_sources WHERE fund_source_uuid = $1", fundSourceUUID)
	var fundSource dbmodels.TreasuryFundSource
	if err := row.Scan(&fundSource.FundSourceUuid, &fundSource.Name, &fundSource.SourceType, &fundSource.Currency, &fundSource.Balance, &fundSource.AvailableBalance, &fundSource.BankInfo, &fundSource.Bin, &fundSource.BankAccountNumber, &fundSource.BankAccountOwner, &fundSource.CashOwner, &fundSource.TenantID, &fundSource.OfficeID); err != nil {
		return dbmodels.TreasuryFundSource{}, err
	}
	return fundSource, nil
}

func walletDomainToExpectedRow(wallet *domain.Wallet) dbmodels.TreasuryWallet {
	return dbmodels.TreasuryWallet{
		WalletUuid: wallet.UUID(),
		Name:       wallet.Name(),
		Currency:   wallet.Balance().Currency(),
		Balance:    wallet.Balance().Amount(),
		TenantID:   "tenant-1",
		OfficeID:   "office-1",
	}
}
