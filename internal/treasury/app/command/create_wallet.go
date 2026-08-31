package command

import (
	"context"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/shopspring/decimal"
)

type CreateWallet struct {
	Name          string
	Currency      shared.Currency
	Allocations   []CreateWalletAllocation
	TenantContext shared.TenantContext
}

type CreateWalletAllocation struct {
	FundSourceUUID   domain.FundSourceUUID
	AllocatedBalance decimal.Decimal
}

func (h *Handlers) CreateWallet(ctx context.Context, cmd CreateWallet) (domain.WalletUUID, error) {
	fsUUIDsForAllocations := make([]domain.FundSourceUUID, 0, len(cmd.Allocations))
	for _, allocation := range cmd.Allocations {
		fsUUIDsForAllocations = append(fsUUIDsForAllocations, allocation.FundSourceUUID)
	}

	var wallet *domain.Wallet
	err := h.walletRepository.CreateWallet(
		ctx,
		cmd.TenantContext,
		fsUUIDsForAllocations,
		func(fundSources map[domain.FundSourceUUID]*domain.FundSource) (*domain.Wallet, error) {
			var err error
			wallet, err = domain.NewWallet(
				cmd.Name,
				cmd.Currency,
			)
			if err != nil {
				return nil, err
			}

			for _, allocation := range cmd.Allocations {
				allocationAmount, err := shared.NewMoney(allocation.AllocatedBalance, cmd.Currency)
				if err != nil {
					return nil, common.NewInvalidInputError(
						"invalid-allocation-amount",
						"%s", err.Error(),
					)
				}
				if err = wallet.Allocate(
					fundSources[allocation.FundSourceUUID],
					allocationAmount,
				); err != nil {
					return nil, common.NewInvalidInputError(
						"failed-to-allocate",
						"%s", err.Error(),
					)
				}

			}

			return wallet, nil
		},
	)
	if err != nil {
		return domain.WalletUUID{}, err
	}

	return wallet.UUID(), nil
}
