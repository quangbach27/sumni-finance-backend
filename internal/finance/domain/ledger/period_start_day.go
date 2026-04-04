package ledger

import "errors"

type PeriodStartDay struct {
	value int32
}

func NewPeriodStartDay(startDay int32) (PeriodStartDay, error) {
	if startDay < 1 || startDay > 28 {
		return PeriodStartDay{}, errors.New("start day must be between 1 and 28")
	}

	return PeriodStartDay{value: startDay}, nil
}

func (p PeriodStartDay) Value() int32 { return p.value }
