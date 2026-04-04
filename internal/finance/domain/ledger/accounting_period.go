package ledger

import (
	"errors"
	"fmt"
	"sumni-finance-backend/internal/common/validator"
	"sumni-finance-backend/internal/common/valueobject"
	"time"

	"github.com/google/uuid"
)

type AccountingPeriod struct {
	id        uuid.UUID
	yearMonth YearMonth

	startDate PeriodStartDay
	interval  int32

	status AccountingPeriodStatus

	openingBalance valueobject.Money
	totalDebit     valueobject.Money
	totalCredit    valueobject.Money
	closingBalance valueobject.Money

	endDate time.Time

	version int32

	transactions []*TransactionRecord
}

func OpenAccountingPeriod(
	yearMonth YearMonth,
	openBalance valueobject.Money,
	startDate PeriodStartDay,
	interval int32,
) (*AccountingPeriod, error) {
	v := validator.New()

	v.Check(!openBalance.IsZero(), "openingBalance", "openingBalance is required")
	v.Check(!yearMonth.IsZero(), "yearMonth", "yearMonth is required")
	if err := v.Err(); err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	zeroMoney, err := valueobject.NewMoney(0, openBalance.Currency())
	if err != nil {
		return nil, err
	}

	endDate := time.Date(
		yearMonth.year,
		time.Month(yearMonth.month),
		int(startDate.value),
		0,
		0,
		0,
		0,
		time.Local,
	).AddDate(0, int(interval), 0)

	return &AccountingPeriod{
		id:             id,
		yearMonth:      yearMonth,
		startDate:      startDate,
		interval:       interval,
		status:         AccountingPeriodOpen,
		openingBalance: openBalance,
		totalDebit:     zeroMoney,
		totalCredit:    zeroMoney,
		closingBalance: zeroMoney,
		endDate:        endDate,
		version:        0,
	}, nil
}

func (ap *AccountingPeriod) IsClose() bool { return ap.status == AccountingPeriodClose }

func (ap *AccountingPeriod) CloseAccountingPeriod() error {
	if time.Now().Before(ap.endDate) {
		return errors.New("too early to close Account Period")
	}

	if ap.IsClose() {
		return fmt.Errorf("account period is already closed: %d/%d", ap.yearMonth.month, ap.yearMonth.year)
	}

	closingBalance, err := ap.calculateClosingBalance()
	if err != nil {
		return err
	}

	ap.closingBalance = closingBalance
	ap.status = AccountingPeriodClose

	return nil
}

func (ap *AccountingPeriod) calculateClosingBalance() (valueobject.Money, error) {
	totalChanged, err := ap.totalCredit.Subtract(ap.totalDebit)
	if err != nil {
		return valueobject.Money{}, err
	}

	closingBalance, err := ap.openingBalance.Add(totalChanged)
	if err != nil {
		return valueobject.Money{}, err
	}

	return closingBalance, nil
}

func (ap *AccountingPeriod) ID() uuid.UUID                      { return ap.id }
func (ap *AccountingPeriod) YearMonth() YearMonth               { return ap.yearMonth }
func (ap *AccountingPeriod) StartDate() PeriodStartDay          { return ap.startDate }
func (ap *AccountingPeriod) Interval() int32                    { return ap.interval }
func (ap *AccountingPeriod) Status() AccountingPeriodStatus     { return ap.status }
func (ap *AccountingPeriod) OpeningBalance() valueobject.Money  { return ap.openingBalance }
func (ap *AccountingPeriod) TotalDebit() valueobject.Money      { return ap.totalDebit }
func (ap *AccountingPeriod) TotalCredit() valueobject.Money     { return ap.totalCredit }
func (ap *AccountingPeriod) ClosingBalance() valueobject.Money  { return ap.closingBalance }
func (ap *AccountingPeriod) EndDate() time.Time                 { return ap.endDate }
func (ap *AccountingPeriod) Version() int32                     { return ap.version }
func (ap *AccountingPeriod) Transactions() []*TransactionRecord { return ap.transactions }

func (ap *AccountingPeriod) Record(txRecord TransactionRecord) error {
	if txRecord.IsCredit() {
		newTotalCredit, err := ap.totalCredit.Add(txRecord.amount)
		if err != nil {
			return err
		}

		ap.totalCredit = newTotalCredit
	} else {
		newTotalDebit, err := ap.totalDebit.Add(txRecord.amount)
		if err != nil {
			return err
		}

		ap.totalDebit = newTotalDebit
	}

	ap.transactions = append(ap.transactions, &txRecord)

	return nil
}

// UnmarshalAccountingPeriodFromDatabase rehydrates an AccountingPeriod from persisted database state.
func UnmarshalAccountingPeriodFromDatabase(
	id uuid.UUID,
	yearMonthStr string,
	startDay int32,
	interval int32,
	statusStr string,
	openingBalanceAmount int64,
	totalDebitAmount int64,
	totalCreditAmount int64,
	closingBalanceAmount int64,
	currencyCode string,
	endDate time.Time,
	version int32,
	transactions ...*TransactionRecord,
) (*AccountingPeriod, error) {
	v := validator.New()

	v.Check(id != uuid.Nil, "id", "id is required")
	v.Required(yearMonthStr, "yearMonth")
	v.Check(interval > 0, "interval", "interval must be greater than 0")
	v.Required(statusStr, "status")
	v.Required(currencyCode, "currencyCode")
	v.Check(!endDate.IsZero(), "endDate", "endDate is required")

	if err := v.Err(); err != nil {
		return nil, err
	}

	yearMonth, err := UnmarshalYearMonthFromString(yearMonthStr)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yearMonth: %w", err)
	}

	periodStartDay, err := NewPeriodStartDay(startDay)
	if err != nil {
		return nil, fmt.Errorf("failed to create PeriodStartDay: %w", err)
	}

	status, err := NewAccountingPeriodStatus(statusStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create AccountingPeriodStatus: %w", err)
	}

	currency, err := valueobject.NewCurrency(currencyCode)
	if err != nil {
		return nil, fmt.Errorf("failed to create Currency: %w", err)
	}

	openingBalance, err := valueobject.NewMoney(openingBalanceAmount, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create opening balance: %w", err)
	}

	totalDebit, err := valueobject.NewMoney(totalDebitAmount, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create total debit: %w", err)
	}

	totalCredit, err := valueobject.NewMoney(totalCreditAmount, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create total credit: %w", err)
	}

	closingBalance, err := valueobject.NewMoney(closingBalanceAmount, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create closing balance: %w", err)
	}

	if transactions == nil {
		transactions = []*TransactionRecord{}
	}

	return &AccountingPeriod{
		id:             id,
		yearMonth:      yearMonth,
		startDate:      periodStartDay,
		interval:       interval,
		status:         status,
		openingBalance: openingBalance,
		totalDebit:     totalDebit,
		totalCredit:    totalCredit,
		closingBalance: closingBalance,
		endDate:        endDate,
		version:        version,
		transactions:   transactions,
	}, nil
}
