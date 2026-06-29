package app

import (
	"context"

	"sumni-finance-backend/internal/common/shared"
)

type PolicyEnforcer interface {
	Enforce(ctx context.Context, accessToken string, resource shared.AuthzResource, scope shared.AuthzScope) (bool, error)
}

type Service struct {
	policyEnforcer PolicyEnforcer
}

func NewService(
	policyEnforcer PolicyEnforcer,
) *Service {
	if policyEnforcer == nil {
		panic("policyEnforcer can't be nil")
	}

	return &Service{
		policyEnforcer: policyEnforcer,
	}
}
