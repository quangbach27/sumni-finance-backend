package domain

import (
	"context"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
)

type BankProvider interface {
	FindByBankCode(ctx context.Context, bankCode string) (BankInfo, error)
}

type FundSourceFactory struct {
	bankProvider BankProvider
}

func NewFundSourceFactory(bankProvider BankProvider) *FundSourceFactory {
	return &FundSourceFactory{bankProvider: bankProvider}
}

func (f *FundSourceFactory) NewFundSource(
	name string,
	sourceType FundSourceType,
	balance shared.Money,
	currency shared.Currency,
	metadata FundSourceMetadata,
	createdBy string,
) (*FundSource, error) {
	errDetails := []common.ErrorDetails{}

	if name == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "empty-name",
			Message:    "name can't be empty",
		})
	}
	if sourceType.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "empty-source-type",
			Message:    "source type can't be empty",
		})
	}
	if currency.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "empty-currency",
			Message:    "currency can't be empty",
		})
	}
	if balance.Amount().IsNegative() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "negative-balance",
			Message:    "balance can't be negative",
		})
	}
	if metadata == nil {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "empty-metadata",
			Message:    "metadata can't be empty",
		})
	} else if !metadata.MatchesType(sourceType) {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSource",
			ErrorSlug:  "metadata-type-mismatch",
			Message:    "metadata type does not match source type",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-fund-source", "fund source is not valid").WithDetails(errDetails)
	}
	return &FundSource{
		uuid:             FundSourceUUID{UUID: common.NewUUIDv7()},
		name:             name,
		sourceType:       sourceType,
		balance:          balance,
		availableBalance: balance,
		currency:         currency,
		metadata:         metadata,
		Audit:            shared.NewAudit(createdBy),
	}, nil
}

func (f *FundSourceFactory) NewBankMetadata(
	ctx context.Context,
	bankCode string,
	accountNumber string,
	accountOwner string,
) (FundSourceBankMetadata, error) {
	bankInfo, err := f.bankProvider.FindByBankCode(ctx, bankCode)
	if err != nil {
		return FundSourceBankMetadata{}, fmt.Errorf("error retrieving bank information: %w", err)
	}

	errDetails := []common.ErrorDetails{}
	if bankInfo.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSourceBankMetadata",
			ErrorSlug:  "empty-bank-info",
			Message:    "bank info can't be empty",
		})
	}
	if accountNumber == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSourceBankMetadata",
			ErrorSlug:  "empty-account-number",
			Message:    "account number can't be empty",
		})
	}
	if accountOwner == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSourceBankMetadata",
			ErrorSlug:  "empty-account-owner",
			Message:    "account owner can't be empty",
		})
	}

	if len(errDetails) != 0 {
		return FundSourceBankMetadata{}, common.NewInvalidInputError("invalid-bank-metadata", "bank metadata is not valid").WithDetails(errDetails)
	}

	return FundSourceBankMetadata{
		bankInfo:      bankInfo,
		accountNumber: accountNumber,
		accountOwner:  accountOwner,
	}, nil
}

func (f *FundSourceFactory) NewCashMetadata(ownerName string) (FundSourceCashMetadata, error) {
	errDetails := []common.ErrorDetails{}
	if ownerName == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "FundSourceCashMetadata",
			ErrorSlug:  "empty-owner-name",
			Message:    "owner name can't be empty",
		})
	}

	if len(errDetails) != 0 {
		return FundSourceCashMetadata{}, common.NewInvalidInputError("invalid-cash-metadata", "cash metadata is not valid").WithDetails(errDetails)
	}

	return FundSourceCashMetadata{ownerName: ownerName}, nil
}
