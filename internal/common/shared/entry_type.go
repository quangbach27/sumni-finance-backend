package shared

import "sumni-finance-backend/internal/common"

var (
	EntryTypeDebit  = common.MustEnum[EntryType]("DEBIT")
	EntryTypeCredit = common.MustEnum[EntryType]("CREDIT")
)

type EntryType struct {
	common.Enum[EntryTypeValues]
}

func (e EntryType) Reverse() EntryType {
	if e.Equal(EntryTypeDebit.Enum) {
		return EntryTypeCredit
	}

	return EntryTypeDebit
}

type EntryTypeValues string

func (e EntryTypeValues) Values() []string {
	return []string{"DEBIT", "CREDIT"}
}
