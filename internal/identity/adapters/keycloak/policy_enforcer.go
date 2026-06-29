package keycloak

import (
	"context"
	"errors"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/log"
	"sumni-finance-backend/internal/common/shared"

	"github.com/Nerzal/gocloak/v14"
)

var (
	ErrUnauthorizedSession = errors.New("token is invalid or expired")
	ErrServiceUnavailable  = errors.New("upstream identity server unreachable")
)

type CheckPermissionPayload struct {
	AccessToken string
	Resource    shared.AuthzResource
	Scope       shared.AuthzScope
}

type PolicyEnforcerConfig struct {
	BaseURL  string
	Realm    string
	ClientID string
}

type PolicyEnforcementPoint struct {
	config PolicyEnforcerConfig
	client gocloak.GoCloakIface
}

func NewPolicyEnforcementPoint(client gocloak.GoCloakIface, config PolicyEnforcerConfig) (*PolicyEnforcementPoint, error) {
	if config.BaseURL == "" {
		return nil, errors.New("baseURL config can't be empty")
	}
	if config.Realm == "" {
		return nil, errors.New("realm config can't be empty")
	}
	if config.ClientID == "" {
		return nil, errors.New("clientID config can't be empty")
	}

	if client == nil {
		return nil, errors.New("gocloak client can't be nil")
	}

	return &PolicyEnforcementPoint{
		config: config,
		client: client,
	}, nil
}

func (p *PolicyEnforcementPoint) Enforce(
	ctx context.Context,
	accessToken string,
	resource shared.AuthzResource,
	scope shared.AuthzScope,
) (bool, error) {
	logger := log.FromContext(ctx).With(
		"resource", resource,
		"scope", scope,
	)

	permissionStr := buildPermission(resource, scope)
	options := gocloak.RequestingPartyTokenOptions{
		GrantType:    common.ToPtr("urn:ietf:params:oauth:grant-type:uma-ticket"),
		Audience:     common.ToPtr(p.config.ClientID),
		ResponseMode: common.ToPtr("permissions"),
		Permissions:  []string{permissionStr},
	}

	permission, err := p.client.GetRequestingPartyPermissionDecision(
		ctx,
		accessToken,
		p.config.Realm,
		options,
	)
	if err != nil {
		var apiErr *gocloak.APIError
		if errors.As(err, &apiErr) {
			if apiErr.Code == 403 {
				logger.Info("Policy enforcement denied access", "error_msg", apiErr.Error())
				return false, nil
			}

			return false, fmt.Errorf("internal keycloak error (code: %d, type: %s): %w", apiErr.Code, apiErr.Type, err)
		}

		return false, fmt.Errorf("failed to connect to auth server: %w", err)
	}

	return common.SafeDeref(permission.Result, false), nil
}

func buildPermission(resource shared.AuthzResource, scope shared.AuthzScope) string {
	return fmt.Sprintf("%s#%s", resource.String(), scope.String())
}
