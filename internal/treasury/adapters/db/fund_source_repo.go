package db

import (
	"context"
	"errors"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/adapters/db/dbmodels"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fundSourceRepo struct {
	db *pgxpool.Pool
}

func NewFundSourceRepository(pgxDb *pgxpool.Pool) *fundSourceRepo {
	if pgxDb == nil {
		panic("pgxDb can't be nil")
	}

	return &fundSourceRepo{
		db: pgxDb,
	}
}

func (r *fundSourceRepo) CreateFundSource(ctx context.Context, tenantContext shared.TenantContext, fundSource *domain.FundSource) error {
	if fundSource == nil {
		return errors.New("fund source can't be empty")
	}

	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		params := dbmodels.InsertFundSourceParams{
			FundSourceUuid:   fundSource.UUID(),
			Name:             fundSource.Name(),
			SourceType:       fundSource.SourceType(),
			Balance:          fundSource.Balance().Amount(),
			AvailableBalance: fundSource.AvailableBalance().Amount(),
			Currency:         fundSource.Currency(),
			TenantID:         tenantContext.TenantID(),
			OfficeID:         tenantContext.OfficeID(),
		}
		addFundSourceMetadataToParams(fundSource, &params)

		err := queries.InsertFundSource(ctx, params)
		if err != nil {
			return fmt.Errorf("error saving fund source: %w", err)
		}

		return nil
	})
}

func addFundSourceMetadataToParams(fs *domain.FundSource, params *dbmodels.InsertFundSourceParams) {
	if bankMetadata, ok := fs.BankMetadata(); ok {
		params.BankInfo = bankMetadata.BankInfo()
		params.Bin = common.ToPtr(bankMetadata.BankInfo().Bin())
		params.BankAccountNumber = common.ToPtr(bankMetadata.AccountNumber())
		params.BankAccountOwner = common.ToPtr(bankMetadata.AccountOwner())
	}

	if cashMetadata, ok := fs.CashMetadata(); ok {
		params.CashOwner = common.ToPtr(cashMetadata.OwnerName())
	}
}
