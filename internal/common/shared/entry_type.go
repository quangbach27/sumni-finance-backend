package shared

import "sumni-finance-backend/internal/common"

var (
	EntryTypeIn  = common.MustEnum[EntryType]("in")
	EntryTypeOut = common.MustEnum[EntryType]("out")
)

type EntryType struct {
	common.Enum[EntryTypeValues]
}

func (e EntryType) Reverse() EntryType {
	if e.Equal(EntryTypeIn.Enum) {
		return EntryTypeOut
	}

	return EntryTypeIn
}

type EntryTypeValues string

func (e EntryTypeValues) Values() []string {
	return []string{"in", "out"}
}
