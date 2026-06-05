package command

import (
	"context"
	"fmt"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/shopspring/decimal"
)

type RegisterFundSourceCmd struct {
	Name            string
	SourceType      domain.FundSourceType
	InitBalance     decimal.Decimal
	Currency        shared.Currency
	BankMetadata    BankMetadataCmd
	CashMetadataCmd CashMetadataCmd
}

type BankMetadataCmd struct {
	BankCode      string
	AccountNumber string
	AccountOwner  string
}

type CashMetadataCmd struct {
	OwnerName string
}

func (h *Handler) RegisterFundSource(ctx context.Context, cmd RegisterFundSourceCmd) (domain.FundSourceUUID, error) {
	metadata, err := h.buildFundSourceMetadata(ctx, cmd)
	if err != nil {
		return domain.FundSourceUUID{}, err
	}

	initBalance, err := shared.NewMoney(cmd.InitBalance, cmd.Currency)
	if err != nil {
		return domain.FundSourceUUID{}, err
	}

	fundSource, err := h.fundSourceFactory.NewFundSource(
		cmd.Name,
		cmd.SourceType,
		initBalance,
		cmd.Currency,
		metadata,
	)
	if err != nil {
		return domain.FundSourceUUID{}, err
	}

	err = h.fundSourceRepository.SaveFundSource(ctx, fundSource)
	if err != nil {
		return domain.FundSourceUUID{}, err
	}

	return fundSource.UUID(), nil
}

func (h *Handler) buildFundSourceMetadata(
	ctx context.Context,
	cmd RegisterFundSourceCmd,
) (domain.FundSourceMetadata, error) {
	switch cmd.SourceType {
	case domain.FundSourceTypeBank:
		return h.fundSourceFactory.NewBankMetadata(
			ctx,
			cmd.BankMetadata.BankCode,
			cmd.BankMetadata.AccountNumber,
			cmd.BankMetadata.AccountOwner,
		)
	case domain.FundSourceTypeCash:
		return h.fundSourceFactory.NewCashMetadata(
			cmd.CashMetadataCmd.OwnerName,
		)
	default:
		return nil, fmt.Errorf(
			"unsupported fund source type: %s",
			cmd.SourceType.String(),
		)
	}
}
