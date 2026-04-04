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

func NewAllocateFundHandler(walletRepo wallet.Repository, fundProviderRepo fundprovider.Repository) AllocateFundHandler {
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

	fpIDs := h.extractUniqueFpIDs(cmd)

	fpLookup, err := h.getFundProvidersByIDs(ctx, fpIDs)
	if err != nil {
		return httperr.NewUnknowError(err, "failed-to-retrieve-fund-provider-lookup")
	}

	logger.Info("retrieved fund providers")

	if err = h.walletRepo.CreateAllocations(
		ctx,
		cmd.WalletID,
		wallet.NewProviderMatchesAnySpec(fpIDs),
		func(w *wallet.Wallet) error {
			for _, ap := range cmd.AllocationProviders {
				fp := fpLookup[ap.ID]
				if fp == nil {
					return fmt.Errorf("fund provider not found: %s", ap.ID.String())
				}

				err = w.AllocateFundProvider(fp, ap.AllocatedAmount)
				if err != nil {
					return err
				}
			}

			return nil
		},
	); err != nil {
		return httperr.NewUnknowError(err, "failed-to-allocate-fund")
	}

	logger.Info("allocated fund provider success")

	return nil
}

func (h *allocateFundHandler) extractUniqueFpIDs(cmd AllocateFundCmd) []uuid.UUID {
	uniqueIDs := make([]uuid.UUID, 0, len(cmd.AllocationProviders))
	seen := make(map[uuid.UUID]struct{}, len(cmd.AllocationProviders))

	for _, p := range cmd.AllocationProviders {
		if _, ok := seen[p.ID]; ok {
			continue
		}
		seen[p.ID] = struct{}{}
		uniqueIDs = append(uniqueIDs, p.ID)
	}

	return uniqueIDs
}

func (h *allocateFundHandler) getFundProvidersByIDs(ctx context.Context, fpIDs []uuid.UUID) (map[uuid.UUID]*fundprovider.FundProvider, error) {
	fps, err := h.fundProviderRepo.GetByIDs(ctx, fpIDs)
	if err != nil {
		return nil, err
	}

	if len(fps) != len(fpIDs) {
		return nil, fmt.Errorf("one or more fund providers were not found for requested IDs: %v", fpIDs)
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
