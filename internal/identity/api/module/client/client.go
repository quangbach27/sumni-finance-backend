package client

import (
	"context"

	"sumni-finance-backend/internal/common/shared"
)

type Identity interface {
	EnforcePolicy(ctx context.Context, req EnforcePolicyRequest) (EnforcePolicyResponse, error)
}

type EnforcePolicyRequest struct {
	AccessToken string
	Resource    shared.AuthzResource
	Scope       shared.AuthzScope
}

type EnforcePolicyResponse struct {
	Allowed bool
}
