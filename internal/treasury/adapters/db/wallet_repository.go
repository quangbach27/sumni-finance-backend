package db

import (
	"context"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/adapters/db/dbmodels"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type walletRepo struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) *walletRepo {
	if db == nil {
		panic("db can't be nil")
	}

	return &walletRepo{
		db: db,
	}
}

func (r *walletRepo) CreateWallet(
	ctx context.Context,
	tenantContext shared.TenantContext,
	fundSourceUUIDs []domain.FundSourceUUID,
	createFn func(
		fundSources map[domain.FundSourceUUID]*domain.FundSource,
	) (*domain.Wallet, error),
) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)
		fundSources, err := listFundSourcesByUUIDs(ctx, queries, fundSourceUUIDs, tenantContext)
		if err != nil {
			return fmt.Errorf("error retrieving fund sources for allocation: %w", err)
		}
		wallet, err := createFn(fundSources)
		if err != nil {
			return err
		}

		params := dbmodels.InsertWalletParams{
			WalletUuid: wallet.UUID(),
			Name:       wallet.Name(),
			Currency:   wallet.Balance().Currency(),
			Balance:    wallet.Balance().Amount(),
			TenantID:   tenantContext.TenantID(),
			OfficeID:   tenantContext.OfficeID(),
		}

		err = queries.InsertWallet(ctx, params)
		if err != nil {
			return fmt.Errorf("error saving wallet: %w", err)
		}

		if len(wallet.Allocations()) > 0 {
			newAllocations := make([]*domain.FundSourceAllocation, 0, len(wallet.Allocations()))
			for _, allocation := range wallet.Allocations() {
				newAllocations = append(newAllocations, allocation)
			}
			return createAllocationsAndUpdateFundSourceAvailableBalance(
				ctx,
				queries,
				wallet.UUID(),
				newAllocations,
			)
		}

		return nil
	})
}

func createAllocationsAndUpdateFundSourceAvailableBalance(
	ctx context.Context,
	queries *dbmodels.Queries,
	walletUUID domain.WalletUUID,
	allocations []*domain.FundSourceAllocation,
) error {
	copyParams := make([]dbmodels.CreateWalletAllocationsParams, 0, len(allocations))
	for _, allocation := range allocations {
		fs := allocation.FundSource()
		if fs == nil {
			return fmt.Errorf("fund source %s not found", allocation.FundSource().UUID().String())
		}

		if err := queries.UpdateFundSourceAvailableBalance(ctx, dbmodels.UpdateFundSourceAvailableBalanceParams{
			FundSourceUuid:      allocation.FundSource().UUID(),
			NewAvailableBalance: fs.AvailableBalance().Amount(),
		}); err != nil {
			return fmt.Errorf("error updating fund source available balance: %w", err)
		}

		copyParams = append(copyParams, dbmodels.CreateWalletAllocationsParams{
			WalletUuid:     walletUUID,
			FundSourceUuid: allocation.FundSource().UUID(),
			Balance:        allocation.Balance().Amount(),
		})
	}

	_, err := queries.CreateWalletAllocations(ctx, copyParams)
	if err != nil {
		return fmt.Errorf("error saving wallet allocations: %w", err)
	}

	return nil
}

func listFundSourcesByUUIDs(
	ctx context.Context,
	queries *dbmodels.Queries,
	fundSourceUUIDs []domain.FundSourceUUID,
	tenantContext shared.TenantContext,
) (map[domain.FundSourceUUID]*domain.FundSource, error) {
	if len(fundSourceUUIDs) == 0 {
		return nil, nil
	}

	params := dbmodels.GetFundSourcesByUUIDsParams{
		FundSourceUuids: make([]common.UUID, 0, len(fundSourceUUIDs)),
		TenantID:        tenantContext.TenantID(),
		OfficeID:        tenantContext.OfficeID(),
	}
	for _, fundSourceUUID := range fundSourceUUIDs {
		params.FundSourceUuids = append(params.FundSourceUuids, common.UUID(fundSourceUUID.UUID))
	}

	fsRows, err := queries.GetFundSourcesByUUIDs(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(fsRows) != len(fundSourceUUIDs) {
		return nil, fmt.Errorf("expected %d fund sources for allocation, got %d", len(fundSourceUUIDs), len(fsRows))
	}

	fundSources := make(map[domain.FundSourceUUID]*domain.FundSource, len(fsRows))
	for _, fsRow := range fsRows {
		fundSources[fsRow.FundSourceUuid] = unmarshalFundSourceRowToDomain(fsRow)
	}

	return fundSources, nil
}

func (r *walletRepo) GetWallet(ctx context.Context, tenant shared.TenantContext, walletUUID domain.WalletUUID) (*domain.Wallet, error) {
	return getWallet(ctx, dbmodels.New(r.db), tenant, walletUUID)
}

func getWallet(
	ctx context.Context,
	queries *dbmodels.Queries,
	tenant shared.TenantContext,
	walletUUID domain.WalletUUID,
) (*domain.Wallet, error) {
	walletRow, err := queries.GetWalletByUUID(ctx, dbmodels.GetWalletByUUIDParams{
		TenantID:   tenant.TenantID(),
		OfficeID:   tenant.OfficeID(),
		WalletUuid: walletUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving wallet with alllocation: %w", err)
	}

	allocationRows, err := queries.GetAllocationsByWalletUUID(ctx, walletUUID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving allocations: %w", err)
	}

	allocations := make(map[domain.FundSourceUUID]*domain.FundSourceAllocation, len(allocationRows))
	for _, allocationRow := range allocationRows {
		fs := unmarshalFundSourceRowToDomain(allocationRow.TreasuryFundSource)
		allocations[fs.UUID()] = domain.UnmarshalFundSourceAllocation(
			fs,
			shared.UnmarshalMoney(
				allocationRow.TreasuryWalletAllocation.Balance,
				walletRow.Currency,
			),
		)
	}

	return unmarshalWalletRowToDomain(walletRow, allocations), nil
}

func (r *walletRepo) LinkFundSources(
	ctx context.Context,
	tenantContext shared.TenantContext,
	walletUUID domain.WalletUUID,
	allocationsData []domain.FundSourceAllocationData,
) error {
	if len(allocationsData) == 0 {
		return common.NewInvalidInputError("", "fund sources for allocations data is missing")
	}

	return common.UpdateInTx(
		ctx,
		r.db,
		func(ctx context.Context, tx pgx.Tx) error {
			queries := dbmodels.New(tx)
			wallet, err := getWallet(ctx, queries, tenantContext, walletUUID)
			if err != nil {
				return fmt.Errorf("error retrieving wallet: %w", err)
			}

			fundSourcesUUID := make([]domain.FundSourceUUID, 0, len(allocationsData))
			for _, allocation := range allocationsData {
				fundSourcesUUID = append(fundSourcesUUID, allocation.FundSourceUUID)
			}

			fundSources, err := listFundSourcesByUUIDs(ctx, queries, fundSourcesUUID, tenantContext)
			if err != nil {
				return fmt.Errorf("error retrieving fund sources: %w", err)
			}

			for _, allocation := range allocationsData {
				fundSource, ok := fundSources[allocation.FundSourceUUID]
				if !ok {
					return fmt.Errorf("fund source %s not found", allocation.FundSourceUUID.String())
				}

				allocatedMoney, err := shared.NewMoney(allocation.AllocatedAmount, wallet.Balance().Currency())
				if err != nil {
					return err
				}

				if err = wallet.Allocate(fundSource, allocatedMoney); err != nil {
					return err
				}
			}

			if err = queries.UpdateWalletBalance(ctx, dbmodels.UpdateWalletBalanceParams{
				Balance:    wallet.Balance().Amount(),
				WalletUuid: walletUUID,
			}); err != nil {
				return fmt.Errorf("error update wallet balance: %w", err)
			}

			allocations := make([]*domain.FundSourceAllocation, 0, len(wallet.Allocations()))
			for _, allocation := range allocationsData {
				walletAllocation, ok := wallet.Allocations()[allocation.FundSourceUUID]
				if !ok {
					return fmt.Errorf("wallet allocation for fund source %s not found", allocation.FundSourceUUID.String())
				}
				allocations = append(allocations, domain.UnmarshalFundSourceAllocation(
					walletAllocation.FundSource(),
					walletAllocation.Balance(),
				))
			}
			if err = createAllocationsAndUpdateFundSourceAvailableBalance(ctx, queries, walletUUID, allocations); err != nil {
				return err
			}

			return nil
		},
	)
}

func unmarshalWalletRowToDomain(walletRow dbmodels.TreasuryWallet, allocations map[domain.FundSourceUUID]*domain.FundSourceAllocation) *domain.Wallet {
	return domain.UnmarshalWallet(
		walletRow.WalletUuid,
		walletRow.Name,
		shared.UnmarshalMoney(walletRow.Balance, walletRow.Currency),
		allocations,
	)
}

func unmarshalFundSourceRowToDomain(fsRow dbmodels.TreasuryFundSource) *domain.FundSource {
	var metadata domain.FundSourceMetadata
	switch fsRow.SourceType {
	case domain.FundSourceTypeBank:
		metadata = domain.UnmarshalFundSourceBankMetadata(
			fsRow.BankInfo,
			*fsRow.BankAccountNumber,
			*fsRow.BankAccountOwner,
		)
	case domain.FundSourceTypeCash:
		metadata = domain.UnmarshalFundSourceCashMetadata(
			*fsRow.CashOwner,
		)
	}

	return domain.UnmarshalFundSource(
		fsRow.FundSourceUuid,
		fsRow.Name,
		fsRow.SourceType,
		shared.UnmarshalMoney(fsRow.Balance, fsRow.Currency),
		shared.UnmarshalMoney(fsRow.AvailableBalance, fsRow.Currency),
		fsRow.Currency,
		metadata,
	)
}
