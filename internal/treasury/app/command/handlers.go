package command

import (
	"sumni-finance-backend/internal/treasury/domain"
)

type Handlers struct {
	fundSourceRepository domain.FundSourceRepository
	fundSourceFactory    *domain.FundSourceFactory
}

func NewHandlers(
	fundSourceRepository domain.FundSourceRepository,
	fundSourceFactory *domain.FundSourceFactory,
) *Handlers {
	if fundSourceRepository == nil {
		panic("fund source repository can't be nil")
	}

	if fundSourceFactory == nil {
		panic("fund source factory can't be nil")
	}

	return &Handlers{
		fundSourceRepository: fundSourceRepository,
		fundSourceFactory:    fundSourceFactory,
	}
}
