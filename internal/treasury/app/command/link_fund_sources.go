package command

import (
	"context"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/shopspring/decimal"
)

type LinkFundSources struct {
	WalletUUID  domain.WalletUUID
	Allocations []LinkFundSourcesAllocation
	Tenant      shared.TenantContext
}

type LinkFundSourcesAllocation struct {
	FundSourceUUID  domain.FundSourceUUID
	AllocatedAmount decimal.Decimal
}

func (h *Handlers) LinkFundSources(ctx context.Context, cmd LinkFundSources) error {
	errDetails := []common.ErrorDetails{}
	if cmd.WalletUUID.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "LinkFundSourcesCommand",
			ErrorSlug:  "invalid-wallet-uuid",
			Message:    "wallet uuid can't not be empty",
		})
	}
	if len(cmd.Allocations) == 0 {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "LinkFundSourcesCommand",
			ErrorSlug:  "invalid-allocations",
			Message:    "allocations data can't not be empty",
		})
	}
	if len(errDetails) != 0 {
		return common.NewInvalidInputError(
			"invalid-link-fund-sources-command",
			"link fund sources command input is not valid",
		).WithDetails(errDetails)
	}

	allocations := make([]domain.FundSourceAllocationData, 0, len(cmd.Allocations))
	for _, allocation := range cmd.Allocations {
		allocations = append(allocations, domain.FundSourceAllocationData{
			FundSourceUUID:  allocation.FundSourceUUID,
			AllocatedAmount: allocation.AllocatedAmount,
		})
	}

	return h.walletRepository.LinkFundSources(ctx, cmd.Tenant, cmd.WalletUUID, allocations)
}
