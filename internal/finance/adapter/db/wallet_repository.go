package db

import (
	"context"
	"errors"
	"fmt"
	common_db "sumni-finance-backend/internal/common/db"
	"sumni-finance-backend/internal/finance/adapter/db/store"
	"sumni-finance-backend/internal/finance/domain/fundprovider"
	"sumni-finance-backend/internal/finance/domain/wallet"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type walletRepo struct {
	queries            *store.Queries
	transactionManager *common_db.PgxTransactionManager
}

func NewWalletRepo(
	queries *store.Queries,
	transactionManager *common_db.PgxTransactionManager,
) (*walletRepo, error) {
	if queries == nil || transactionManager == nil {
		return nil, errors.New("missing dependencies")
	}

	return &walletRepo{
		queries:            queries,
		transactionManager: transactionManager,
	}, nil
}

func (r *walletRepo) GetByID(
	ctx context.Context,
	wID uuid.UUID,
) (*wallet.Wallet, error) {
	return r.getByID(ctx, wID, r.queries)
}

func (r *walletRepo) getByID(
	ctx context.Context,
	wID uuid.UUID,
	queries *store.Queries,
) (*wallet.Wallet, error) {
	wModel, err := queries.GetWalletByID(ctx, wID)
	if err != nil {
		return nil, err
	}

	w, err := wallet.UnmarshalWalletFromDatabase(
		wModel.ID,
		wModel.Name,
		wModel.Balance,
		wModel.Currency,
		wModel.Version,
	)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (r *walletRepo) GetByIDWithProviders(
	ctx context.Context,
	wID uuid.UUID,
	spec wallet.ProviderAllocationSpec,
) (*wallet.Wallet, error) {
	return r.getByIDWithProviders(
		ctx,
		wID,
		spec,
		r.queries,
	)
}

func (r *walletRepo) getByIDWithProviders(
	ctx context.Context,
	wID uuid.UUID,
	spec wallet.ProviderAllocationSpec,
	queries *store.Queries,
) (*wallet.Wallet, error) {
	wModel, err := queries.GetWalletByID(ctx, wID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve wallet '%s': %w", wID.String(), err)
	}

	fpModels, err := queries.GetFundProviderByWalletID(ctx, wID)
	if err != nil {
		return nil, err
	}

	filteredAllocations := make([]wallet.ProviderAllocation, 0, len(fpModels))
	for _, fpModel := range fpModels {
		fp, err := fundprovider.UnmarshalFundProviderFromDatabase(
			fpModel.ID,
			fpModel.Name,
			fpModel.FpType,
			fpModel.Balance,
			fpModel.UnallocatedAmount,
			fpModel.Currency,
			fpModel.Version,
		)
		if err != nil {
			return nil, err
		}

		allocation, err := wallet.NewProviderAllocation(fp, fpModel.WalletAllocatedAmount)
		if err != nil {
			return nil, err
		}

		if spec == nil {
			filteredAllocations = append(filteredAllocations, allocation)
		} else if spec.IsSatisfiedBy(allocation) {
			filteredAllocations = append(filteredAllocations, allocation)
		}
	}

	w, err := wallet.UnmarshalWalletFromDatabase(
		wModel.ID,
		wModel.Name,
		wModel.Balance,
		wModel.Currency,
		wModel.Version,
		filteredAllocations...,
	)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (r *walletRepo) Create(ctx context.Context, wallet *wallet.Wallet) error {
	return r.queries.CreateWallet(ctx, store.CreateWalletParams{
		ID:       wallet.ID(),
		Name:     wallet.Name(),
		Balance:  wallet.Balance().Amount(),
		Currency: wallet.Currency().Code(),
		Version:  0,
	})
}

func (r *walletRepo) CreateAllocations(
	ctx context.Context,
	wID uuid.UUID,
	allocationSpec wallet.ProviderAllocationSpec,
	updateFunc func(*wallet.Wallet) error,
) error {
	return r.transactionManager.WithTx(ctx, func(tx pgx.Tx) error {
		txQueries := r.queries.WithTx(tx)

		w, err := r.getByIDWithProviders(
			ctx,
			wID,
			allocationSpec,
			txQueries,
		)
		if err != nil {
			return err
		}

		if err = updateFunc(w); err != nil {
			return err
		}

		rows, err := txQueries.UpdateWalletBalance(ctx, store.UpdateWalletBalanceParams{
			ID:      w.ID(),
			Balance: w.Balance().Amount(),
			Version: w.Version(),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("failed to update wallet balance: %w", common_db.ErrConcurrentModification)
		}

		return r.insertFundAllocations(ctx, txQueries, w.ID(), w.ProviderManager().ProviderAllocations())
	})
}

func (r *walletRepo) insertFundAllocations(
	ctx context.Context,
	queries *store.Queries,
	wID uuid.UUID,
	fpAllocations []wallet.ProviderAllocation,
) error {
	allocationsLen := len(fpAllocations)
	if allocationsLen == 0 {
		return errors.New("empty fund provider allocation")
	}

	allocationParams := make([]store.BulkInsertFundAllocationsParams, 0, allocationsLen)
	fpParams := store.BatchUpdateFundProvidersBalanceParams{
		Ids:                make([]uuid.UUID, 0, allocationsLen),
		Balances:           make([]int64, 0, allocationsLen),
		UnallocatedAmounts: make([]int64, 0, allocationsLen),
		Versions:           make([]int32, 0, allocationsLen),
	}

	for _, fpa := range fpAllocations {
		fp := fpa.Provider()
		if fp == nil {
			return errors.New("fund provider is missing")
		}

		allocationParams = append(allocationParams, store.BulkInsertFundAllocationsParams{
			FpID:            fp.ID(),
			WalletID:        wID,
			AllocatedAmount: fpa.Allocated().Amount(),
		})

		fpParams.Ids = append(fpParams.Ids, fp.ID())
		fpParams.Balances = append(fpParams.Balances, fp.Balance().Amount())
		fpParams.UnallocatedAmounts = append(fpParams.UnallocatedAmounts, fp.UnallocatedBalance().Amount())
		fpParams.Versions = append(fpParams.Versions, fp.Version())
	}

	rows, err := queries.BatchUpdateFundProvidersBalance(ctx, fpParams)
	if err != nil {
		return err
	}
	if rows != int64(allocationsLen) {
		return fmt.Errorf("failed to update fund provider when allocations: %w", common_db.ErrConcurrentModification)
	}

	row, err := queries.BulkInsertFundAllocations(ctx, allocationParams)
	if err != nil {
		return err
	}

	if row != int64(len(fpAllocations)) {
		return errors.New("bulk updated fund allocations failed")
	}

	return nil
}
