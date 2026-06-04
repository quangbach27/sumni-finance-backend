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
	UpdateFundSource(
		ctx context.Context,
		fundSourceUUID FundSourceUUID,
		updateFn func(fs *FundSource) error,
	)
	RecordJournalEntries(
		ctx context.Context,
		fundSourceUUID FundSourceUUID,
		recordFn func(fundSource *FundSource) ([]*JournalEntry, error),
	) error
	VoidJournalEntry(
		ctx context.Context,
		fundSourceUUID FundSourceUUID,
		journalEntryUUID JournalEntryUUID,
		updateFn func(foundSource *FundSource, journalEntryForVoided *JournalEntry) (*JournalEntry, error),
	) (JournalEntryUUID, error)
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
	shared.Audit
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

func (fp *FundSource) CashMetadata() (FundSourceCashMetadata, bool) {
	m, ok := fp.metadata.(FundSourceCashMetadata)
	return m, ok
}

func (fp *FundSource) TopUp(m shared.Money) error {
	if m.Amount().IsZero() {
		return errors.New("money for top up can't be empty")
	}

	if m.Amount().IsNegative() {
		return errors.New("money for top up can't be negative")
	}

	newBalance, err := fp.balance.Add(m)
	if err != nil {
		return fmt.Errorf("failed to topup: %w", err)
	}

	fp.balance = newBalance

	return nil
}

func (fp *FundSource) Withdraw(m shared.Money) error {
	if m.Amount().IsZero() {
		return errors.New("money for withdraw can't be empty")
	}

	if m.Amount().IsNegative() {
		return errors.New("money for withdraw can't be negative")
	}

	if fp.balance.Amount().LessThan(m.Amount()) {
		return errors.New("withdraw amount can't exceed current balance")
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
		return errors.New("allocation money currency must match fund provider currency")
	}

	if m.Amount().IsNegative() {
		return errors.New("allocation money can't be negative")
	}

	if fp.availableBalance.Amount().LessThan(m.Amount()) {
		return errors.New("allocation money can't exceed fund source available balance")
	}

	newAvailableBalance, err := fp.availableBalance.Sub(m)
	if err != nil {
		return err
	}

	fp.availableBalance = newAvailableBalance

	return nil
}
