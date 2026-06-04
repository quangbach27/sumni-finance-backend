package shared

import (
	"time"

	"github.com/shopspring/decimal"
)

func UnmarshalMoney(
	amount decimal.Decimal,
	currency Currency,
) Money {
	return Money{amount: amount, currency: currency}
}

func UnmarshalAudit(
	createdAt time.Time,
	createdBy string,
	updatedAt *time.Time,
	updatedBy *string,
) Audit {
	return Audit{
		createdAt: createdAt,
		createdBy: createdBy,
		updatedAt: updatedAt,
		updatedBy: updatedBy,
	}
}
