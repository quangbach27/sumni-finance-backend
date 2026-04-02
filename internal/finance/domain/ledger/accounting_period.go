package ledger

import (
	"errors"
	"fmt"
	"sumni-finance-backend/internal/common/validator"
	"sumni-finance-backend/internal/common/valueobject"
	"time"
)

type AccountingPeriod struct {
	yearMonth YearMonth

	startDate PeriodStartDay
	interval  int

	status AccountingPeriodStatus

	openingBalance valueobject.Money
	totalDebit     valueobject.Money
	totalCredit    valueobject.Money
	closingBalance valueobject.Money

	endDate time.Time

	transactions []*TransactionRecord
}

func OpenAccountingPeriod(
	yearMonth YearMonth,
	openBalance valueobject.Money,
	startDate PeriodStartDay,
	interval int,
) (*AccountingPeriod, error) {
	v := validator.New()

	v.Check(!openBalance.IsZero(), "openingBalance", "openingBalance is required")
	v.Check(!yearMonth.IsZero(), "yearMonth", "yearMonth is required")
	if err := v.Err(); err != nil {
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
	).AddDate(0, interval, 0)

	return &AccountingPeriod{
		yearMonth:      yearMonth,
		startDate:      startDate,
		interval:       interval,
		status:         AccountingPeriodOpen,
		openingBalance: openBalance,
		totalDebit:     zeroMoney,
		totalCredit:    zeroMoney,
		closingBalance: zeroMoney,
		endDate:        endDate,
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

func (ap *AccountingPeriod) YearMonth() YearMonth               { return ap.yearMonth }
func (ap *AccountingPeriod) StartDate() PeriodStartDay          { return ap.startDate }
func (ap *AccountingPeriod) Interval() int                      { return ap.interval }
func (ap *AccountingPeriod) Status() AccountingPeriodStatus     { return ap.status }
func (ap *AccountingPeriod) OpeningBalance() valueobject.Money  { return ap.openingBalance }
func (ap *AccountingPeriod) TotalDebit() valueobject.Money      { return ap.totalDebit }
func (ap *AccountingPeriod) TotalCredit() valueobject.Money     { return ap.totalCredit }
func (ap *AccountingPeriod) ClosingBalance() valueobject.Money  { return ap.closingBalance }
func (ap *AccountingPeriod) EndDate() time.Time                 { return ap.endDate }
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
