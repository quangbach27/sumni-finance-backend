package models

import "fmt"

type ClaimsSource interface {
	Claims(v any) error
}

type Organization struct {
	id     string
	name   string
	groups []string
}

func (o Organization) ID() string       { return o.id }
func (o Organization) Name() string     { return o.name }
func (o Organization) Groups() []string { return o.groups }

type TokenClaims struct {
	subject      string
	username     string
	email        string
	realmRoles   []string            // Global roles
	clientRoles  map[string][]string // Client-specific roles map
	organization Organization
}

func (t TokenClaims) Subject() string                  { return t.subject }
func (t TokenClaims) Username() string                 { return t.username }
func (t TokenClaims) Email() string                    { return t.email }
func (t TokenClaims) RealmRoles() []string             { return t.realmRoles }
func (t TokenClaims) ClientRoles() map[string][]string { return t.clientRoles }
func (t TokenClaims) Organization() Organization       { return t.organization }

type keycloakClaims struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"` // Captures client roles mapped by clientID
	Organization map[string]struct {
		ID     string   `json:"id"`
		Groups []string `json:"groups"`
	} `json:"organization"`
}

func UnmarshalTokenClaims(source ClaimsSource) (TokenClaims, error) {
	var raw keycloakClaims
	if err := source.Claims(&raw); err != nil {
		return TokenClaims{}, fmt.Errorf("failed to unmarshal source claims: %w", err)
	}

	var organization Organization

	for orgName, orgData := range raw.Organization {
		organization = Organization{
			id:     orgData.ID,
			name:   orgName,
			groups: orgData.Groups,
		}

		break
	}

	clientRoles := make(map[string][]string, len(raw.ResourceAccess))
	for clientID, access := range raw.ResourceAccess {
		clientRoles[clientID] = access.Roles
	}

	return TokenClaims{
		subject:      raw.Sub,
		username:     raw.PreferredUsername,
		email:        raw.Email,
		realmRoles:   raw.RealmAccess.Roles,
		clientRoles:  clientRoles,
		organization: organization,
	}, nil
}
