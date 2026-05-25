package shared

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type Money struct {
	amount   decimal.Decimal
	currency Currency
}

func NewMoney(amount decimal.Decimal, currency Currency) (Money, error) {
	if currency.IsZero() {
		return Money{}, errors.New("currency can't be empty")
	}

	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

func (m Money) Amount() decimal.Decimal { return m.amount }
func (m Money) Currency() Currency      { return m.currency }

type ErrCurrencyMismatch struct {
	Left  Currency
	Rigth Currency
}

func (e ErrCurrencyMismatch) Error() string {
	return fmt.Sprintf("currency mismatch: source %s, destination %s",
		e.Left.String(), e.Rigth.String())
}

func (m Money) Add(other Money) (Money, error) {
	if !m.currency.Equal(other.currency) {
		return Money{}, ErrCurrencyMismatch{
			Left:  m.currency,
			Rigth: other.currency,
		}
	}

	return Money{
		amount:   m.amount.Add(other.amount),
		currency: m.currency,
	}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if !m.currency.Equal(other.currency) {
		return Money{}, ErrCurrencyMismatch{
			Left:  m.currency,
			Rigth: other.currency,
		}
	}

	return Money{
		amount:   m.amount.Sub(other.amount),
		currency: m.currency,
	}, nil
}

func (m Money) IsZero() bool {
	return m == Money{}
}

func (m Money) Equal(o Money) bool {
	return m.amount.Equal(o.amount) && m.currency.Equal(o.currency)
}

func (m Money) String() string {
	return fmt.Sprintf("%s-%s", m.amount.String(), m.currency.String())
}

type moneyDto struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency Currency        `json:"currency"`
}

func (m *Money) UnmarshalJSON(data []byte) error {
	var dto moneyDto
	err := json.Unmarshal(data, &dto)
	if err != nil {
		return fmt.Errorf("failed to unmarshal money: %w", err)
	}

	m.amount = dto.Amount
	m.currency = dto.Currency

	return nil
}

func (m Money) MarshalJSON() ([]byte, error) {
	dto := moneyDto{
		Amount:   m.amount,
		Currency: m.currency,
	}

	return json.Marshal(dto)
}
