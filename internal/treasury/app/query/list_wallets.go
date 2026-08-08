package query

import (
	"context"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/shopspring/decimal"
)

type WalletReadModel interface {
	ListWallets(ctx context.Context, query ListWallets) (ListWalletsReadModel, error)
}

type ListWallets struct {
	Page          int
	PageSize      int
	TenantContext shared.TenantContext
}

type WalletAllocationItem struct {
	FundSourceUUID domain.FundSourceUUID
	FundSourceName string
	Balance        decimal.Decimal
}

type ListWalletItem struct {
	UUID        domain.WalletUUID
	Name        string
	Currency    shared.Currency
	Balance     decimal.Decimal
	Allocations []WalletAllocationItem
}

type ListWalletsReadModel common.Pagination[ListWalletItem]

func (h *Handlers) ListWallets(ctx context.Context, query ListWallets) (ListWalletsReadModel, error) {
	if query.TenantContext.IsZero() {
		return ListWalletsReadModel{}, common.NewInvalidInputError("empty-tenant", "tenant can't not be empty")
	}

	return h.walletReadModel.ListWallets(ctx, query)
}
