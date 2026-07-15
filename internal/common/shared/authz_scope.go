package shared

import "sumni-finance-backend/internal/common"

type AuthzScope struct {
	common.Enum[AuthzScopeType]
}

type AuthzScopeType string

func (v AuthzScopeType) Values() []string {
	return []string{"read", "write", "delete"}
}

var (
	AuthzScopeRead   = common.MustEnum[AuthzScope]("read")
	AuthzScopeWrite  = common.MustEnum[AuthzScope]("write")
	AuthzScopeDelete = common.MustEnum[AuthzScope]("delete")
)
