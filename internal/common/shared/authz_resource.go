package shared

import "sumni-finance-backend/internal/common"

type AuthzResource struct {
	common.Enum[AuthzResourceType]
}

type AuthzResourceType string

func (t AuthzResourceType) Values() []string {
	return []string{"fund-source", "journal-entry"}
}

var (
	AuthzResourceFundSource   = common.MustEnum[AuthzResource]("fund-source")
	AuthzResourceJournalEntry = common.MustEnum[AuthzResource]("journal-entry")
)
