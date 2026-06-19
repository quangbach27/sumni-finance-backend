package domain

import (
	"context"
	"errors"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
)

type FundSourceRepository interface {
	SaveFundSource(ctx context.Context, fundSource *FundSource) error
	RecordJournalEntries(
		ctx context.Context,
		fundSourceUUID FundSourceUUID,
		recordFn func(fundSource *FundSource) ([]*JournalEntry, error),
	) error
	VoidJournalEntry(
		ctx context.Context,
		fundSourceUUID FundSourceUUID,
		journalEntryUUIDToVoid JournalEntryUUID,
		voidFn func(
			fundSource *FundSource,
			journalEntryToVoid *JournalEntry,
		) (*JournalEntry, error),
	) (*JournalEntry, error)
}

type FundSourceUUID struct {
	common.UUID
}

type FundSourceType struct {
	common.Enum[FundSourceTypeValues]
}

type FundSourceTypeValues string

func (ts FundSourceTypeValues) Values() []string {
	return []string{"BANK", "CASH"}
}

var (
	FundSourceTypeBank = common.MustEnum[FundSourceType]("BANK")
	FundSourceTypeCash = common.MustEnum[FundSourceType]("CASH")
)

type FundSource struct {
	uuid             FundSourceUUID
	name             string
	sourceType       FundSourceType
	balance          shared.Money
	availableBalance shared.Money
	currency         shared.Currency

	metadata FundSourceMetadata
	audit    *shared.Audit
}

func (f *FundSource) UUID() FundSourceUUID           { return f.uuid }
func (f *FundSource) Name() string                   { return f.name }
func (f *FundSource) SourceType() FundSourceType     { return f.sourceType }
func (f *FundSource) Balance() shared.Money          { return f.balance }
func (f *FundSource) AvailableBalance() shared.Money { return f.availableBalance }
func (f *FundSource) Currency() shared.Currency      { return f.currency }
func (f *FundSource) Metadata() FundSourceMetadata   { return f.metadata }
func (fp *FundSource) BankMetadata() (FundSourceBankMetadata, bool) {
	m, ok := fp.metadata.(FundSourceBankMetadata)
	return m, ok
}
func (fp *FundSource) Audit() *shared.Audit { return fp.audit }

func (fp *FundSource) CashMetadata() (FundSourceCashMetadata, bool) {
	m, ok := fp.metadata.(FundSourceCashMetadata)
	return m, ok
}

func (fp *FundSource) TopUp(m shared.Money) error {
	if !m.Amount().IsPositive() {
		return errors.New("top up amount must be positive")
	}

	newBalance, err := fp.balance.Add(m)
	if err != nil {
		return fmt.Errorf("failed to topup: %w", err)
	}

	fp.balance = newBalance

	return nil
}

func (fp *FundSource) Withdraw(m shared.Money) error {
	if !m.Amount().IsPositive() {
		return errors.New("withdraw amount must be positive")
	}

	if fp.balance.Amount().LessThan(m.Amount()) {
		return errors.New("insufficient balance to complete withdrawal")
	}

	newBalance, err := fp.balance.Sub(m)
	if err != nil {
		return fmt.Errorf("failed to withdraw: %w", err)
	}

	fp.balance = newBalance

	return nil
}

func (fp *FundSource) Reserve(m shared.Money) error {
	if !fp.currency.Equal(m.Currency()) {
		return fmt.Errorf("reservation currency %s does not match fund source currency %s", m.Currency(), fp.currency)
	}

	if m.Amount().IsNegative() {
		return errors.New("reservation amount must be greater or equal than zero")
	}

	if fp.availableBalance.Amount().LessThan(m.Amount()) {
		return errors.New("insufficient available balance to complete reservation")
	}

	newAvailableBalance, err := fp.availableBalance.Sub(m)
	if err != nil {
		return err
	}

	fp.availableBalance = newAvailableBalance

	return nil
}
