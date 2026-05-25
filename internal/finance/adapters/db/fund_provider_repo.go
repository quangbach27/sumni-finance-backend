package db

import (
	"context"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/finance/adapters/db/dbmodels"
	"sumni-finance-backend/internal/finance/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FundProviderRepository struct {
	db *pgxpool.Pool
}

func NewFundProviderRepository(db *pgxpool.Pool) *FundProviderRepository {
	if db == nil {
		panic("db can't be nil")
	}

	return &FundProviderRepository{
		db: db,
	}
}

func (r *FundProviderRepository) GetFundProviders(
	ctx context.Context,
	fundProviderUUID []domain.FundProviderUUID,
) ([]*domain.FundProvider, error) {
	return nil, nil
}

func (r *FundProviderRepository) SaveFundProvider(
	ctx context.Context,
	fundProvider *domain.FundProvider,
) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		params := dbmodels.InsertFundProviderParams{
			FundProviderUuid: fundProvider.UUID(),
			Name:             fundProvider.Name(),
			Balance:          fundProvider.Balance().Amount(),
			AvailableBalance: fundProvider.AvailableMoney().Amount(),
			FundProviderType: fundProvider.Type(),
			Currency:         fundProvider.Currency(),
			OfficeUuid:       fundProvider.OfficeUUID(),
		}
		addFundProviderMetadataToParams(fundProvider, &params)

		err := queries.InsertFundProvider(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to insert fund provider: %w", err)
		}

		return nil
	})
}

func addFundProviderMetadataToParams(fp *domain.FundProvider, params *dbmodels.InsertFundProviderParams) {
	if bankMetadata, ok := fp.BankMetadata(); ok {
		params.BankAccountNumber = common.Ptr(bankMetadata.AccountNumber())
		params.BankName = common.Ptr(bankMetadata.BankName())
		params.BankAccountOwner = common.Ptr(bankMetadata.AccountOwner())
	}

	if cashMetadata, ok := fp.CashMetadata(); ok {
		params.CashOwnerName = common.Ptr(cashMetadata.OwnerName())
	}
}
