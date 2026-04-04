package db

import (
	"context"
	"sumni-finance-backend/internal/common/convert"
	"sumni-finance-backend/internal/finance/adapter/db/store"
	"sumni-finance-backend/internal/finance/domain/ledger"

	"github.com/google/uuid"
)

type ledgerRepository struct {
	queries *store.Queries
}

func NewLedgerRepository(queries *store.Queries) *ledgerRepository {
	return &ledgerRepository{
		queries: queries,
	}
}

func (r *ledgerRepository) CreateAccountingPeriod(
	ctx context.Context,
	wID uuid.UUID,
	ap *ledger.AccountingPeriod,
) error {
	return r.queries.CreateAccountingPeriod(ctx, store.CreateAccountingPeriodParams{
		ID:                   ap.ID(),
		YearMonth:            ap.YearMonth().String(),
		StartDate:            ap.StartDate().Value(),
		Interval:             ap.Interval(),
		EndTime:              ap.EndDate(),
		WalletOpeningBalance: ap.OpeningBalance().Amount(),
		TotalDebit:           ap.TotalDebit().Amount(),
		TotalCredit:          ap.TotalCredit().Amount(),
		WalletClosingBalance: ap.ClosingBalance().Amount(),
		Version:              convert.SafePtr(ap.Version()),
		Status:               ap.Status().String(),
		WalletID:             wID,
	})
}
