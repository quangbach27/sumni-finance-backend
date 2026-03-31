package ledger

import "errors"

type PeriodStartDay struct {
	value uint8
}

func NewPeriodStartDay(startDay uint8) (PeriodStartDay, error) {
	if startDay < 1 || startDay > 28 {
		return PeriodStartDay{}, errors.New("start day must be between 1 and 28")
	}

	return PeriodStartDay{value: startDay}, nil
}

func (p PeriodStartDay) Value() uint8 { return p.value }
