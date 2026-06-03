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
			AvailableBalance: fundProvider.AvailableBalance().Amount(),
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
		params.BankInfo = bankMetadata.BankInfo()
		params.BankAccountNumber = common.ToPtr(bankMetadata.AccountNumber())
		params.BankAccountOwner = common.ToPtr(bankMetadata.AccountOwner())
	}

	if cashMetadata, ok := fp.CashMetadata(); ok {
		params.CashOwnerName = common.ToPtr(cashMetadata.OwnerName())
	}
}

func unmarshalFundProviderFromDB(fundProviderDb dbmodels.FinancesFundProvider) *domain.FundProvider {
	return domain.UnmarshalFundProvider(
		fundProviderDb.FundProviderUuid,
		fundProviderDb.OfficeUuid,
		fundProviderDb.Name,
		fundProviderDb.FundProviderType,
		shared.UnmarshalMoney(fundProviderDb.Balance, fundProviderDb.Currency),
		shared.UnmarshalMoney(fundProviderDb.AvailableBalance, fundProviderDb.Currency),
		fundProviderDb.Currency,
		unmarshalFundProviderMetadata(fundProviderDb),
	)
}

func unmarshalFundProviderMetadata(fpModel dbmodels.FinancesFundProvider) domain.FundProviderMetadata {
	var metadata domain.FundProviderMetadata

	switch fpModel.FundProviderType {
	case domain.FundProviderTypeBank:
		metadata = domain.UnmarshalFundProviderBankMetadata(
			fpModel.BankInfo,
			common.Deref(fpModel.BankAccountNumber, ""),
			common.Deref(fpModel.BankAccountOwner, ""),
		)
	case domain.FundProviderTypeCash:
		metadata = domain.UnmarshalFundProviderCashMetadata(
			common.Deref(fpModel.CashOwnerName, ""),
		)
	}

	return metadata
}
