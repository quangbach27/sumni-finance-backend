package query

import "context"

func (h *Handlers) ListFundSources(ctx context.Context) ([]FundSourceView, error) {
	return h.fundSourceReadModel.ListFundSources(ctx)
}
