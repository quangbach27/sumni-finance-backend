package fundprovider

import (
	"errors"
	"strings"
)

var ErrInvalidType = errors.New("invalid fund provider type")
var (
	CashType Type = Type{value: "CASH"}
	BankType Type = Type{value: "BANK"}
)

var supportedType = map[string]Type{
	"CASH": CashType,
	"BANK": BankType,
}

type Type struct {
	value string
}

func NewType(typeStr string) (Type, error) {
	typeCleaned := strings.TrimSpace(strings.ToUpper(typeStr))

	t, ok := supportedType[typeCleaned]
	if !ok {
		return Type{}, ErrInvalidType
	}

	return t, nil
}

func (t Type) String() string {
	return t.value
}

func (t Type) IsZero() bool {
	return t == Type{}
}
