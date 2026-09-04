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
) BankMetadata {
	return BankMetadata{
		bankInfo:      bankInfo,
		accountNumber: accountNumber,
		accountOwner:  accountOwner,
	}
}

func UnmarshalFundSourceCashMetadata(
	ownerName string,
) CashMetadata {
	return CashMetadata{
		ownerName: ownerName,
	}
}

func UnmarshalWallet(
	uuid WalletUUID,
	name string,
	balance shared.Money,
	allocations map[FundSourceUUID]*FundSourceAllocation,
) *Wallet {
	return &Wallet{
		uuid:        uuid,
		name:        name,
		balance:     balance,
		allocations: allocations,
	}
}

func UnmarshalFundSourceAllocation(
	fs *FundSource,
	balance shared.Money,
) *FundSourceAllocation {
	return &FundSourceAllocation{
		fundSource: fs,
		balance:    balance,
	}
}
