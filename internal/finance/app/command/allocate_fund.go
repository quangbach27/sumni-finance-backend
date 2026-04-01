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
	WalletID            uuid.UUID
	AllocationProviders []AllocatedProvider
}

type AllocatedProvider struct {
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

	if len(cmd.AllocationProviders) == 0 {
		return httperr.NewIncorrectInputError(
			errors.New("missing providers in command for allocation"),
			"invalid-providers",
		)
	}

	fpIDs := make([]uuid.UUID, 0, len(cmd.AllocationProviders))
	for _, p := range cmd.AllocationProviders {
		fpIDs = append(fpIDs, p.ID)
	}

	fpLookup, err := h.getFundProvidersByIDs(ctx, fpIDs)
	if err != nil {
		return httperr.NewIncorrectInputError(err, "failed-to-retrieve-fund-provider-lookup")
	}

	logger.Info("retrived fund providers")

	err = h.walletRepo.CreateAllocations(
		ctx,
		cmd.WalletID,
		func(w *wallet.Wallet) error {
			for _, ap := range cmd.AllocationProviders {
				fp := fpLookup[ap.ID]

				err = w.AllocateFromFundProvider(fp, ap.AllocatedAmount)
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

	logger.Info("allocated fund provider success")

	return nil
}

func (h *allocateFundHandler) getFundProvidersByIDs(ctx context.Context, fpIDs []uuid.UUID) (map[uuid.UUID]*fundprovider.FundProvider, error) {
	fps, err := h.fundProviderRepo.GetByIDs(ctx, fpIDs)
	if err != nil {
		return nil, err
	}

	if len(fps) != len(fpIDs) {
		return nil, errors.New("")
	}

	return h.toFundProviderLookup(fps)
}

func (h *allocateFundHandler) toFundProviderLookup(fps []*fundprovider.FundProvider) (map[uuid.UUID]*fundprovider.FundProvider, error) {
	fpLookup := make(map[uuid.UUID]*fundprovider.FundProvider, len(fps))
	for _, fp := range fps {
		if fp == nil {
			return nil, fmt.Errorf("fund provider is empty")
		}

		fpLookup[fp.ID()] = fp
	}

	return fpLookup, nil
}
