package ledger

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type YearMonth struct {
	year  int
	month int
}

func NewYearMonth(
	month,
	year int,
) (YearMonth, error) {
	if year < 1 {
		return YearMonth{}, errors.New("year must be greater than 0")
	}

	if month < 1 || month > 12 {
		return YearMonth{}, errors.New("month must be between 1 and 12")
	}

	return YearMonth{
		year:  year,
		month: month,
	}, nil
}

func UnmarshalYearMonthFromDatabase(ymStr string) (YearMonth, error) {
	ymStrCleaned := strings.TrimSpace(ymStr)

	res := strings.Split(ymStrCleaned, ",")
	if len(res) != 2 {
		return YearMonth{}, fmt.Errorf("unknown ymStr format: %s", ymStr)
	}

	year, err := strconv.Atoi(res[0])
	if err != nil {
		return YearMonth{}, fmt.Errorf("failed to parse year: %w", err)
	}

	month, err := strconv.Atoi(res[1])
	if err != nil {
		return YearMonth{}, fmt.Errorf("failed to parse month: %w", err)
	}

	ym, err := NewYearMonth(month, year)
	if err != nil {
		return YearMonth{}, err
	}

	return ym, nil
}

func (ym YearMonth) Year() int      { return ym.year }
func (ym YearMonth) Month() int     { return ym.month }
func (ym YearMonth) IsZero() bool   { return ym == YearMonth{} }
func (ym YearMonth) String() string { return fmt.Sprintf("%d,%d", ym.year, ym.month) }
