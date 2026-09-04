package query

import (
	"context"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/shopspring/decimal"
)

type FundSourceReadModel interface {
	ListFundSources(ctx context.Context, query ListFundSources) (ListFundSourcesReadModel, error)
}

type ListFundSources struct {
	Page          int
	PageSize      int
	TenantContext shared.TenantContext
}

type FundSourceBankMetadata struct {
	BankName      string
	BankBin       string
	BankShortName string
	AccountNumber string
	AccountOwner  string
}

type FundSourceCashMetadata struct {
	OwnerName string
}

type ListFundSourceItem struct {
	UUID             domain.FundSourceUUID
	Name             string
	SourceType       domain.FundSourceType
	Currency         shared.Currency
	Balance          decimal.Decimal
	AvailableBalance decimal.Decimal
	BankMetadata     *FundSourceBankMetadata
	CashMetadata     *FundSourceCashMetadata
}

type ListFundSourcesReadModel common.Pagination[ListFundSourceItem]

func (h *Handlers) ListFundSources(ctx context.Context, query ListFundSources) (ListFundSourcesReadModel, error) {
	if query.TenantContext.IsZero() {
		return ListFundSourcesReadModel{}, common.NewInvalidInputError("empty-tenant", "tenant can't not be empty")
	}

	return h.fundSourceReadModel.ListFundSources(ctx, query)
}
