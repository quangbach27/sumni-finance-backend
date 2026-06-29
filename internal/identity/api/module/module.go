package module

import (
	"context"

	"sumni-finance-backend/internal/identity/api/module/client"
	"sumni-finance-backend/internal/identity/app"
)

type Identity struct {
	service *app.Service
}

func New(service *app.Service) *Identity {
	if service == nil {
		panic("service can't be nil")
	}

	return &Identity{
		service: service,
	}
}

func (i *Identity) EnforcePolicy(
	ctx context.Context,
	req client.EnforcePolicyRequest,
) (client.EnforcePolicyResponse, error) {
	allowed, err := i.service.EnforcePolicy(
		ctx,
		app.EnforcePolicyCmd{
			AccessToken: req.AccessToken,
			Resource:    req.Resource,
			Scope:       req.Scope,
		},
	)
	if err != nil {
		return client.EnforcePolicyResponse{}, err
	}

	return client.EnforcePolicyResponse{
		Allowed: allowed,
	}, nil
}
