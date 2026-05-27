package domain

import (
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/finance/app/models"
)

func UnmarshalFundProviderBankMetadata(
	bankName string,
	accountNumber string,
	accountOwner string,
) FundProviderBankMetadata {
	return FundProviderBankMetadata{
		accountNumber: accountNumber,
		accountOwner:  accountOwner,
		bankName:      bankName,
	}
}

func UnmarshalFundProviderCashMetadata(ownerName string) FundProviderCashMetadata {
	return FundProviderCashMetadata{ownerName: ownerName}
}

func UnmarshalFundProvider(
	uuid FundProviderUUID,
	officeUUID models.OfficeUUID,
	name string,
	fundProviderType FundProviderType,
	balance shared.Money,
	availableBalance shared.Money,
	currency shared.Currency,
	metadata FundProviderMetadata,
) *FundProvider {
	return &FundProvider{
		uuid:             uuid,
		officeUUID:       officeUUID,
		name:             name,
		fundProviderType: fundProviderType,
		balance:          balance,
		availableBalance: availableBalance,
		currency:         currency,
		metadata:         metadata,
	}
}

func UnmarshalWallet(
	uuid WalletUUID,
	name string,
	description string,
	balance shared.Money,
	currency shared.Currency,
	officeUUID models.OfficeUUID,
	allocationsData []NewFundProviderAllocationData,
) *Wallet {
	register := &fundProviderRegistry{
		currency:    currency,
		allocations: make(map[FundProviderUUID]*FundProviderAllocation),
	}
	for _, data := range allocationsData {
		register.allocations[data.FundProvider.UUID()] = &FundProviderAllocation{
			fundProvider: data.FundProvider,
			amount:       data.AllocationAmount,
		}
	}

	return &Wallet{
		uuid:        uuid,
		name:        name,
		description: description,
		balance:     balance,
		currency:    currency,
		officeUUID:  officeUUID,
		fpRegistry:  register,
	}
}
