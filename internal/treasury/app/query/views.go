package query

import (
	"time"

	"github.com/shopspring/decimal"

	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"
)

type PaginatedResult[T any] struct {
	Items      []T
	TotalItems int
	Page       int
	PageSize   int
}

type FundSourceView struct {
	UUID             domain.FundSourceUUID
	Name             string
	SourceType       domain.FundSourceType
	Balance          decimal.Decimal
	AvailableBalance decimal.Decimal
	Currency         shared.Currency

	BankMetadata *FundSourceBankMetadataView
	CashMetadata *FundSourceCashMetadataView

	CreatedAt time.Time
	CreatedBy string
	UpdatedAt *time.Time
	UpdatedBy *string
}

type FundSourceBankMetadataView struct {
	BankName      string
	BankCode      string
	BankShortName string
	BankBin       int
	BankLogoUrl   string
	BankIconUrl   string
	AccountNumber string
	AccountOwner  string
}

type FundSourceCashMetadataView struct {
	OwnerName string
}

type JournalEntryView struct {
	UUID            domain.JournalEntryUUID
	Amount          decimal.Decimal
	EntryType       shared.EntryType
	TransactionDate time.Time
	TransactionNo   *string
	Description     *string
	Status          domain.JournalEntryStatus
	FundSourceUUID  domain.FundSourceUUID
	BalanceBefore   decimal.Decimal
	BalanceAfter    decimal.Decimal

	VoidedBy         *string
	VoidedAt         *time.Time
	VoidedReason     *string
	ReverseEntryUUID *domain.JournalEntryUUID

	CreatedAt time.Time
	CreatedBy string
	UpdatedAt *time.Time
	UpdatedBy *string
}
