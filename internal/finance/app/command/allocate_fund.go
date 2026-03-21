package command

import (
	"context"
	"errors"
	"fmt"
	"sumni-finance-backend/internal/common/cqrs"
	"sumni-finance-backend/internal/common/logs"
	"sumni-finance-backend/internal/common/server/httperr"
	"sumni-finance-backend/internal/finance/domain/fundprovider"
	"sumni-finance-backend/internal/finance/domain/wallet"

	"github.com/google/uuid"
)

type AllocateFundCmd struct {
	WalletID  uuid.UUID
	Providers []AllocatedProviders
}

type AllocatedProviders struct {
	ID              uuid.UUID
	AllocatedAmount int64
}

type AllocateFundHandler cqrs.CommandHandler[AllocateFundCmd]

type allocateFundHandler struct {
	walletRepo       wallet.Repository
	fundProviderRepo fundprovider.Repository
}

func NewAllocateFundHandler(walletRepo wallet.Repository, fundProviderRepo fundprovider.Repository) *allocateFundHandler {
	return &allocateFundHandler{
		walletRepo:       walletRepo,
		fundProviderRepo: fundProviderRepo,
	}
}

func (h *allocateFundHandler) Handle(ctx context.Context, cmd AllocateFundCmd) error {
	logger := logs.FromContext(ctx)

	if len(cmd.Providers) == 0 {
		return httperr.NewIncorrectInputError(
			errors.New("missing providers in command for allocation"),
			"invalid-providers",
		)
	}

	providerIDs := make([]uuid.UUID, 0, len(cmd.Providers))
	for _, p := range cmd.Providers {
		providerIDs = append(providerIDs, p.ID)
	}

	// Retrieve all fund providers involved in this allocation request
	providersDomain, err := h.fundProviderRepo.GetByIDs(ctx, providerIDs)
	if err != nil {
		return httperr.NewUnknowError(err, "failed-to-retrieve-fund-provider")
	}
	if len(providersDomain) != len(providerIDs) {
		return httperr.NewIncorrectInputError(
			errors.New("one or more fund providers not found"),
			"invalid-provider",
		)
	}

	logger.Info("retrieved provider", "ids", providerIDs)

	// Build lookup map for deterministic access
	providerMap := make(map[uuid.UUID]*fundprovider.FundProvider, len(providersDomain))
	for _, p := range providersDomain {
		if p == nil {
			return fmt.Errorf("fund provider is empty")
		}
		providerMap[p.ID()] = p
	}

	err = h.walletRepo.Update(
		ctx,
		cmd.WalletID,
		wallet.NewAllocationBelongsToAnyProviderSpec(providerIDs),
		func(w *wallet.Wallet) error {
			for _, item := range cmd.Providers {
				provider, ok := providerMap[item.ID]
				if !ok {
					return fmt.Errorf("fund provider %s not found", item.ID)
				}

				err = w.AllocateFromFundProvider(provider, item.AllocatedAmount)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)
	if err != nil {
		return httperr.NewUnknowError(err, "failed-to-allocate-fund")
	}

	return nil
}
