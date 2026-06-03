package db

import (
	"context"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/finance/adapters/db/dbmodels"
	"sumni-finance-backend/internal/finance/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepo struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) *WalletRepo {
	if db == nil {
		panic("db can't be empty")
	}

	return &WalletRepo{db: db}
}

func (r *WalletRepo) GetWalletWithAllocations(
	ctx context.Context,
	walletUUID domain.WalletUUID,
) (*domain.Wallet, error) {
	return r.getWalletWithAllocations(ctx, walletUUID, r.db)
}

func (r *WalletRepo) getWalletWithAllocations(
	ctx context.Context,
	walletUUID domain.WalletUUID,
	db dbmodels.DBTX,
) (*domain.Wallet, error) {
	queries := dbmodels.New(db)

	walletDb, err := queries.WalletByUUID(ctx, walletUUID)
	if err != nil {
		return nil, fmt.Errorf("faiiled to get wallet: %w", err)
	}

	allocationRows, err := queries.FundProvidersWithAllocations(ctx, walletUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fund provider allocations: %w", err)
	}

	allocationsData := make([]domain.NewFundProviderAllocationData, 0, len(allocationRows))
	for _, allocationRow := range allocationRows {
		allocationsData = append(allocationsData, domain.NewFundProviderAllocationData{
			FundProvider: unmarshalFundProviderFromDB(allocationRow.FinancesFundProvider),
			AllocationAmount: shared.UnmarshalMoney(
				allocationRow.FinancesFundProviderAllocation.AllocationAmount,
				walletDb.Currency,
			),
		})
	}

	return unmarshalWalletFromDB(walletDb, allocationsData), nil
}

func (r *WalletRepo) SaveWallet(
	ctx context.Context,
	fundProvidersUUIDs []domain.FundProviderUUID,
	createFn func(
		fundProviders map[domain.FundProviderUUID]*domain.FundProvider,
	) (*domain.Wallet, error),
) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		fundProviders, err := r.getFundProvidersByUUIDs(ctx, queries, fundProvidersUUIDs)
		if err != nil {
			return err
		}

		wallet, err := createFn(fundProviders)
		if err != nil {
			return fmt.Errorf("error create wallet: %w", err)
		}

		err = queries.InsertWallet(ctx, dbmodels.InsertWalletParams{
			WalletUuid:  wallet.UUID(),
			Name:        wallet.Name(),
			Description: wallet.Description(),
			Balance:     wallet.Balance().Amount(),
			Currency:    wallet.Currency(),
			OfficeUuid:  wallet.OfficeUUID(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert wallet: %w", err)
		}

		for _, allocation := range wallet.FundProviderAllocations() {
			err = queries.UpsertFundProviderAllocations(ctx, dbmodels.UpsertFundProviderAllocationsParams{
				WalletUuid:       wallet.UUID(),
				FundProviderUuid: allocation.FundProvider().UUID(),
				AllocationAmount: allocation.Amount().Amount(),
			})
			if err != nil {
				return fmt.Errorf("failed to upsert fund provider allocation")
			}

			err = queries.UpdateFundProviderBalance(ctx, dbmodels.UpdateFundProviderBalanceParams{
				FundProviderUuid: allocation.FundProvider().UUID(),
				AvailableBalance: common.ToPtr(allocation.FundProvider().AvailableBalance().Amount()),
			})
			if err != nil {
				return fmt.Errorf("failed to update fund provider available balance: %w", err)
			}
		}

		return nil
	})
}

func (r *WalletRepo) UpdateFundProviderAllocations(
	ctx context.Context,
	walletUUID domain.WalletUUID,
	updateFn func(wallet *domain.Wallet) error,
) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		return nil
	})
}

func (r *WalletRepo) getFundProvidersByUUIDs(
	ctx context.Context,
	queries *dbmodels.Queries,
	fundProviderUUIDs []domain.FundProviderUUID,
) (map[domain.FundProviderUUID]*domain.FundProvider, error) {
	if len(fundProviderUUIDs) == 0 {
		return nil, nil
	}

	commonUUIDs := make([]common.UUID, 0, len(fundProviderUUIDs))

	for _, fpUUID := range fundProviderUUIDs {
		commonUUIDs = append(commonUUIDs, fpUUID.UUID)
	}

	fundProvidersDB, err := queries.FundProvidersByUUIDs(ctx, commonUUIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get fund provider by uuids: %w", err)
	}

	if len(fundProvidersDB) != len(commonUUIDs) {
		return nil, common.NewInvalidInputError(
			"invalid-fund-provider-uuids",
			"fund provider uuids is %d but only got %d in database", len(fundProviderUUIDs), len(fundProvidersDB),
		)
	}

	fundProviders := make(map[domain.FundProviderUUID]*domain.FundProvider, len(fundProviderUUIDs))

	for _, fpDB := range fundProvidersDB {
		fundProviders[fpDB.FundProviderUuid] = unmarshalFundProviderFromDB(fpDB)
	}

	return fundProviders, err
}

func unmarshalWalletFromDB(walletDb dbmodels.FinancesWallet, allocationsData []domain.NewFundProviderAllocationData) *domain.Wallet {
	return domain.UnmarshalWallet(
		walletDb.WalletUuid,
		walletDb.Name,
		walletDb.Description,
		shared.UnmarshalMoney(walletDb.Balance, walletDb.Currency),
		walletDb.Currency,
		walletDb.OfficeUuid,
		allocationsData,
	)
}
