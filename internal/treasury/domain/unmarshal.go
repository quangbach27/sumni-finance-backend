package domain

import "sumni-finance-backend/internal/common/shared"

func UnmarshalFundSource(
	uuid FundSourceUUID,
	name string,
	sourceType FundSourceType,
	balance shared.Money,
	availableBalance shared.Money,
	currency shared.Currency,
	metadata FundSourceMetadata,
) *FundSource {
	return &FundSource{
		uuid:             uuid,
		name:             name,
		sourceType:       sourceType,
		balance:          balance,
		availableBalance: availableBalance,
		currency:         currency,
		metadata:         metadata,
	}
}

func UnmarshalFundSourceBankMetadata(
	bankInfo BankInfo,
	accountNumber string,
	accountOwner string,
) FundSourceBankMetadata {
	return FundSourceBankMetadata{
		bankInfo:      bankInfo,
		accountNumber: accountNumber,
		accountOwner:  accountOwner,
	}
}

func UnmarshalFundSourceCashMetadata(
	ownerName string,
) FundSourceCashMetadata {
	return FundSourceCashMetadata{
		ownerName: ownerName,
	}
}
