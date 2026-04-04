package ledger

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateAccountingPeriod(
		ctx context.Context,
		wID uuid.UUID,
		ap *AccountingPeriod,
	) error
}
