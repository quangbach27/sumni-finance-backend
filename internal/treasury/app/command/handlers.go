package command

import "sumni-finance-backend/internal/treasury/domain"

type Handlers struct {
	walletRepository     domain.WalletRepository
	fundSourceRepository domain.FundSourceRepository
}

func NewHandlers(
	walletRepository domain.WalletRepository,
	fundSourceRepository domain.FundSourceRepository,
) *Handlers {
	if walletRepository == nil {
		panic("walletRepository can't be nil")
	}

	if fundSourceRepository == nil {
		panic("fundSourceRepository can't be nil")
	}

	return &Handlers{
		walletRepository:     walletRepository,
		fundSourceRepository: fundSourceRepository,
	}
}
