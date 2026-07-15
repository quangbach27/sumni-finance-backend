package app

import (
	"context"

	"sumni-finance-backend/internal/common/shared"
)

type EnforcePolicyCmd struct {
	AccessToken string
	Resource    shared.AuthzResource
	Scope       shared.AuthzScope
}

func (s *Service) EnforcePolicy(ctx context.Context, cmd EnforcePolicyCmd) (bool, error) {
	return s.policyEnforcer.Enforce(ctx, cmd.AccessToken, cmd.Resource, cmd.Scope)
}
