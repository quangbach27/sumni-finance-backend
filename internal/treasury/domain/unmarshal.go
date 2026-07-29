package domain

import (
	"time"

	"sumni-finance-backend/internal/common/shared"

	"github.com/shopspring/decimal"
)

func UnmarshalFundSource(
	uuid FundSourceUUID,
	name string,
	sourceType FundSourceType,
	balance shared.Money,
	availableBalance shared.Money,
	currency shared.Currency,
	metadata FundSourceMetadata,
	audit shared.Audit,
) *FundSource {
	return &FundSource{
		uuid:             uuid,
		name:             name,
		sourceType:       sourceType,
		balance:          balance,
		availableBalance: availableBalance,
		currency:         currency,
		metadata:         metadata,
		Audit:            audit,
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

func UnmarshalJournalEntry(
	uuid JournalEntryUUID,
	amount decimal.Decimal,
	entryType shared.EntryType,
	transactionDate time.Time,
	transactionNo *string,
	description *string,
	status JournalEntryStatus,
	fundSourceUUID FundSourceUUID,
	balanceBefore decimal.Decimal,
	balanceAfter decimal.Decimal,
	voidedBy *string,
	voidedAt *time.Time,
	voidedReason *string,
	reverseEntryUUID *JournalEntryUUID,
	audit shared.Audit,
) *JournalEntry {
	return &JournalEntry{
		uuid:             uuid,
		amount:           amount,
		entryType:        entryType,
		transactionDate:  transactionDate,
		transactionNo:    transactionNo,
		description:      description,
		status:           status,
		fundSourceUUID:   fundSourceUUID,
		balanceBefore:    balanceBefore,
		balanceAfter:     balanceAfter,
		voidedBy:         voidedBy,
		voidedAt:         voidedAt,
		voidedReason:     voidedReason,
		reverseEntryUUID: reverseEntryUUID,
		Audit:            audit,
	}
}
