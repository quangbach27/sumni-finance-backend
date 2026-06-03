package domain

import (
	"errors"
	"time"

	"sumni-finance-backend/internal/common"
)

type LedgerAccountUUID struct {
	common.UUID
}

type PeriodFrequency struct {
	common.Enum[PeriodFrequencyValues]
}

type PeriodFrequencyValues string

func (PeriodFrequencyValues) Values() []string {
	return []string{"MONTHLY", "QUARTERLY", "ANNUALLY"}
}

var (
	PeriodFrequencyMonthly   = common.MustEnum[PeriodFrequency]("MONTHLY")
	PeriodFrequencyQuarterly = common.MustEnum[PeriodFrequency]("QUARTERLY")
	PeriodFrequencyAnnually  = common.MustEnum[PeriodFrequency]("ANNUALLY")
)

type LedgerPeriodConfig struct {
	dayOfMonth int
	frequency  PeriodFrequency
}

func NewLedgerPeriodConfig(dayOfMonth int, frequency PeriodFrequency) (LedgerPeriodConfig, error) {
	if dayOfMonth < 1 || dayOfMonth > 28 {
		return LedgerPeriodConfig{}, errors.New("dayOfMonth must be between 1 and 28")
	}

	if frequency.IsZero() {
		return LedgerPeriodConfig{}, errors.New("frequency can't be empty")
	}

	return LedgerPeriodConfig{
		dayOfMonth: dayOfMonth,
		frequency:  frequency,
	}, nil
}

func (c LedgerPeriodConfig) Frequency() PeriodFrequency {
	return c.frequency
}

func (c LedgerPeriodConfig) DayOfMonth() int {
	return c.dayOfMonth
}

func (c LedgerPeriodConfig) IsZero() bool {
	return c == LedgerPeriodConfig{}
}

func (c LedgerPeriodConfig) CalculateCloseDate(from time.Time) time.Time {
	switch c.frequency {
	case PeriodFrequencyMonthly:
		return from.AddDate(0, 1, 0)
	case PeriodFrequencyQuarterly:
		return from.AddDate(0, 3, 0)
	case PeriodFrequencyAnnually:
		return from.AddDate(0, 12, 0)
	default:
		return from.AddDate(0, 1, 0)
	}
}
