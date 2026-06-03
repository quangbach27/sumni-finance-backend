package domain

import (
	"errors"
	"fmt"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"

	"github.com/shopspring/decimal"
)

type LedgerPeriodUUID struct {
	common.UUID
}

type LedgerPeriod struct {
	uuid        LedgerPeriodUUID
	accountUUID LedgerAccountUUID
	config      LedgerPeriodConfig

	openBalance  decimal.Decimal
	totalDebit   decimal.Decimal
	totalCredit  decimal.Decimal
	closeBalance decimal.Decimal

	reported bool
	closed   bool

	startDate time.Time
	closeDate time.Time

	entries []LedgerEntry
}

func NewLedgerPeriod(
	accountUUID LedgerAccountUUID,
	config LedgerPeriodConfig,
	openBalance decimal.Decimal,
) (*LedgerPeriod, error) {
	if accountUUID.IsZero() {
		return nil, errors.New("account uuid can't be empty")
	}
	if config.IsZero() {
		return nil, errors.New("ledger period config can't be empty")
	}

	now := time.Now().UTC()
	startDate := time.Date(now.Year(), now.Month(), config.dayOfMonth, 0, 0, 0, 0, time.UTC)

	return &LedgerPeriod{
		uuid:        LedgerPeriodUUID{UUID: common.NewUUIDv7()},
		accountUUID: accountUUID,
		config:      config,
		openBalance: openBalance,
		startDate:   startDate,
		closeDate:   config.CalculateCloseDate(startDate),
	}, nil
}

func NewNextLedgerPeriod(prev *LedgerPeriod, config LedgerPeriodConfig) (*LedgerPeriod, error) {
	if prev == nil {
		return nil, errors.New("previous ledger period can't be empty")
	}
	if config.IsZero() {
		return nil, errors.New("ledger period config can't be empty")
	}
	if !prev.closed {
		return nil, errors.New("current ledger period has not closed yet")
	}

	newStartDate := prev.closeDate.AddDate(0, 0, 1)
	newCloseDate := config.CalculateCloseDate(newStartDate)

	return &LedgerPeriod{
		uuid:        LedgerPeriodUUID{UUID: common.NewUUIDv7()},
		accountUUID: prev.accountUUID,
		config:      config,
		openBalance: prev.closeBalance,
		startDate:   newStartDate,
		closeDate:   newCloseDate,
	}, nil
}

func (lp *LedgerPeriod) Close() error {
	if lp.closed {
		return errors.New("ledger period has already closed")
	}
	if lp.closeDate.Before(time.Now().UTC()) {
		return errors.New("the ledger period is ended yet")
	}

	lp.closed = true
	lp.closeBalance = lp.openBalance.Add(lp.totalCredit.Sub(lp.totalDebit))

	return nil
}

func (lp *LedgerPeriod) PostJournalEntry(journalEntry *JournalEntry, forcePost bool) error {
	if journalEntry == nil {
		return errors.New("journal entry can't be empty")
	}
	if lp.closed {
		return errors.New("ledger period has closed")
	}
	if !forcePost && (journalEntry.transactionDate.Before(lp.startDate) || journalEntry.transactionDate.After(lp.closeDate)) {
		return errors.New("journal entry is not in ledger period. Using force post to skip this validation")
	}

	ledgerEntry, err := NewLedgerEntry(journalEntry.amount, journalEntry.entryType, journalEntry.uuid)
	if err != nil {
		return fmt.Errorf("error creating ledger entry: %w", err)
	}

	switch ledgerEntry.entryType {
	case shared.EntryTypeCredit:
		lp.totalCredit = lp.totalCredit.Add(ledgerEntry.amount)
	case shared.EntryTypeDebit:
		lp.totalDebit = lp.totalDebit.Add(ledgerEntry.amount)
	default:
		return fmt.Errorf("unsupport entry type :%s", ledgerEntry.entryType.String())
	}

	lp.entries = append(lp.entries, ledgerEntry)
	return nil
}

func (lp *LedgerPeriod) UUID() LedgerPeriodUUID         { return lp.uuid }
func (lp *LedgerPeriod) AccountUUID() LedgerAccountUUID { return lp.accountUUID }
func (lp *LedgerPeriod) Config() LedgerPeriodConfig     { return lp.config }
func (lp *LedgerPeriod) OpenBalance() decimal.Decimal   { return lp.openBalance }
func (lp *LedgerPeriod) CloseBalance() decimal.Decimal  { return lp.closeBalance }
func (lp *LedgerPeriod) TotalDebit() decimal.Decimal    { return lp.totalDebit }
func (lp *LedgerPeriod) TotalCredit() decimal.Decimal   { return lp.totalCredit }
func (lp *LedgerPeriod) IsClosed() bool                 { return lp.closed }
func (lp *LedgerPeriod) IsReported() bool               { return lp.reported }
func (lp *LedgerPeriod) StartDate() time.Time           { return lp.startDate }
func (lp *LedgerPeriod) CloseDate() time.Time           { return lp.closeDate }
func (lp *LedgerPeriod) Entries() []LedgerEntry         { return lp.entries }

type LedgerEntry struct {
	amount    decimal.Decimal
	entryType shared.EntryType

	postingReference JournalEntryUUID
}

func NewLedgerEntry(
	amount decimal.Decimal,
	entryType shared.EntryType,
	postingReference JournalEntryUUID,
) (LedgerEntry, error) {
	if amount.IsZero() {
		return LedgerEntry{}, errors.New("amount can't be empty")
	}
	if entryType.IsZero() {
		return LedgerEntry{}, errors.New("entry type can't be empty")
	}
	if postingReference.IsZero() {
		return LedgerEntry{}, errors.New("posting reference can't be empty")
	}

	return LedgerEntry{
		amount:           amount,
		entryType:        entryType,
		postingReference: postingReference,
	}, nil
}

func (le LedgerEntry) Amount() decimal.Decimal {
	return le.amount
}

func (le LedgerEntry) EntryType() shared.EntryType {
	return le.entryType
}

func (le LedgerEntry) PostingReference() JournalEntryUUID {
	return le.postingReference
}

func (le LedgerEntry) IsZero() bool {
	return le == LedgerEntry{}
}
