package query

type Handlers struct {
	fundSourceReadModel FundSourceReadModel
	walletReadModel     WalletReadModel
}

func NewHandlers(
	fundSourceReadModel FundSourceReadModel,
	walletReadModel WalletReadModel,
) *Handlers {
	if fundSourceReadModel == nil {
		panic("fundSourceReadModel can't be nil")
	}
	if walletReadModel == nil {
		panic("walletReadModel can't be nil")
	}

	return &Handlers{
		fundSourceReadModel: fundSourceReadModel,
		walletReadModel:     walletReadModel,
	}
}
