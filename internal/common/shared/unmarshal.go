package shared

import "github.com/shopspring/decimal"

func UnmarshalMoney(
	amount decimal.Decimal,
	currency Currency,
) Money {
	return Money{amount: amount, currency: currency}
}
