package command

import (
	"context"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/shopspring/decimal"
)

type CreateFundSource struct {
	Name            string
	SourceType      domain.FundSourceType
	InitBalance     decimal.Decimal
	Currency        shared.Currency
	BankMetadata    BankMetadata
	CashMetadataCmd CashMetadata
	TenantContext   shared.TenantContext
}

type BankMetadata struct {
	BankName      string
	BankBin       string
	BankShortName string
	AccountNumber string
	AccountOwner  string
}

type CashMetadata struct {
	OwnerName string
}

func (h *Handlers) CreateFundSource(ctx context.Context, cmd CreateFundSource) (domain.FundSourceUUID, error) {
	metadata, err := buildFundSourceMetadata(cmd)
	if err != nil {
		return domain.FundSourceUUID{}, err
	}

	initBalance, err := shared.NewMoney(cmd.InitBalance, cmd.Currency)
	if err != nil {
		return domain.FundSourceUUID{}, common.NewInvalidInputError(
			"invalid-init-balance",
			"%s",
			err.Error(),
		)
	}

	fundSource, err := domain.NewFundSource(
		cmd.Name,
		cmd.SourceType,
		initBalance,
		cmd.Currency,
		metadata,
	)
	if err != nil {
		return domain.FundSourceUUID{}, err
	}

	err = h.fundSourceRepository.CreateFundSource(ctx, cmd.TenantContext, fundSource)
	if err != nil {
		return domain.FundSourceUUID{}, err
	}

	return fundSource.UUID(), nil
}

func buildFundSourceMetadata(cmd CreateFundSource) (domain.FundSourceMetadata, error) {
	switch cmd.SourceType {
	case domain.FundSourceTypeBank:
		return domain.NewBankMetadata(
			cmd.BankMetadata.AccountNumber,
			cmd.BankMetadata.AccountOwner,
			domain.BankInfoData{
				Name:      cmd.BankMetadata.BankName,
				Bin:       cmd.BankMetadata.BankBin,
				ShortName: cmd.BankMetadata.BankShortName,
			},
		)
	case domain.FundSourceTypeCash:
		return domain.NewCashMetadata(cmd.CashMetadataCmd.OwnerName)
	default:
		return nil, common.NewInvalidInputError(
			"fund-source-type-unsupported",
			"unsupported fund source type: %s",
			cmd.SourceType.String(),
		)
	}
}
