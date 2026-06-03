package domain

import (
	"errors"
	"strings"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
)

type FundProviderMetadata interface {
	MatchesType(accountType FundProviderType) bool
	IsZero() bool
}

type FundProviderBankMetadata struct {
	bankInfo shared.BankInfo

	accountNumber string
	accountOwner  string
}

func NewFundProviderBankMetadata(
	bankInfo shared.BankInfo,
	accountNumber string,
	accountOwner string,
) (FundProviderBankMetadata, error) {
	errDetails := []common.ErrorDetails{}

	if bankInfo.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankMetadata",
			ErrorSlug:  "empty-bank-info",
			Message:    "bank info can't be empty",
		})
	}
	if accountNumber == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankMetadata",
			ErrorSlug:  "empty-bank-info",
			Message:    "bank info can't be empty",
		})
	}
	if accountOwner == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "BankMetadata",
			ErrorSlug:  "empty-account-owner",
			Message:    "account owner can't be empty",
		})
	}

	if len(errDetails) != 0 {
		return FundProviderBankMetadata{}, common.NewInvalidInputError("invalid-bank-metadata", "bank metadata is not valid").WithDetails(errDetails)
	}

	return FundProviderBankMetadata{
		bankInfo:      bankInfo,
		accountNumber: accountNumber,
		accountOwner:  accountOwner,
	}, nil
}

func (bm FundProviderBankMetadata) BankInfo() shared.BankInfo {
	return bm.bankInfo
}

func (bm FundProviderBankMetadata) AccountNumber() string {
	return bm.accountNumber
}

func (bm FundProviderBankMetadata) AccountOwner() string {
	return bm.accountOwner
}

func (bm FundProviderBankMetadata) MatchesType(fpType FundProviderType) bool {
	return fpType == FundProviderTypeBank
}

func (bm FundProviderBankMetadata) IsZero() bool {
	return bm == FundProviderBankMetadata{}
}

type FundProviderCashMetadata struct {
	ownerName string
}

func NewCashFundProviderMetadata(ownerName string) (FundProviderCashMetadata, error) {
	cleansedOwnerName := strings.TrimSpace(ownerName)
	if cleansedOwnerName == "" {
		return FundProviderCashMetadata{}, errors.New("owner name can't be empty")
	}

	return FundProviderCashMetadata{ownerName: cleansedOwnerName}, nil
}

func (cm FundProviderCashMetadata) OwnerName() string {
	return cm.ownerName
}

func (cm FundProviderCashMetadata) MatchesType(fpType FundProviderType) bool {
	return fpType == FundProviderTypeCash
}

func (cm FundProviderCashMetadata) IsZero() bool {
	return cm == FundProviderCashMetadata{}
}
