package command

import (
	"context"
	"sumni-finance-backend/internal/common/cqrs"
	"sumni-finance-backend/internal/common/server/httperr"
	"sumni-finance-backend/internal/finance/domain/fundprovider"
)

type CreateFundProviderCmd struct {
	Name         string
	FpType       string
	InitBalance  int64
	CurrencyCode string
}

type CreateFundProviderHandler cqrs.CommandHandler[CreateFundProviderCmd]

type createFundProviderHandler struct {
	fundProviderRepo fundprovider.Repository
}

func NewCreateFundProviderHandler(fundProviderRepo fundprovider.Repository) CreateFundProviderHandler {
	return &createFundProviderHandler{
		fundProviderRepo: fundProviderRepo,
	}
}

func (h *createFundProviderHandler) Handle(ctx context.Context, cmd CreateFundProviderCmd) error {
	fundProvider, err := fundprovider.NewFundProvider(
		cmd.Name,
		cmd.FpType,
		cmd.InitBalance,
		cmd.CurrencyCode,
	)
	if err != nil {
		return httperr.NewIncorrectInputError(err, "invalid-cmd-input")
	}

	err = h.fundProviderRepo.Create(ctx, fundProvider)
	if err != nil {
		return httperr.NewUnknowError(err, "failed-to-create-fund-provider")
	}

	return nil
}
