package shared

import (
	"fmt"

	"sumni-finance-backend/internal/common"
)

type Currency struct {
	common.Enum[CurrencyType]
}

func (c Currency) Code() string {
	return c.String()
}

func (c Currency) Equal(o Currency) bool {
	return c.String() == o.String()
}

type CurrencyType string

func (c CurrencyType) Values() []string {
	return []string{"VND", "KRW"}
}

func MustNewCurrency(value string) Currency {
	c := Currency{}
	err := c.UnmarshalText([]byte(value))
	if err != nil {
		panic(fmt.Errorf("error unmarshalling currency value: %s", value))
	}

	return c
}
