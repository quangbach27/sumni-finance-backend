package command

import (
	"sumni-finance-backend/internal/treasury/domain"
)

type Handler struct {
	fundSourceRepository domain.FundSourceRepository
	fundSourceFactory    *domain.FundSourceFactory
}

func NewHandler(
	fundSourceRepository domain.FundSourceRepository,
	fundSourceFactory *domain.FundSourceFactory,
) *Handler {
	if fundSourceRepository == nil {
		panic("fund source repository can't be nil")
	}

	if fundSourceFactory == nil {
		panic("fund source factory can't be nil")
	}

	return &Handler{
		fundSourceRepository: fundSourceRepository,
		fundSourceFactory:    fundSourceFactory,
	}
}
