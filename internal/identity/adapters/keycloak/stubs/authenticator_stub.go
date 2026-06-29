package stubs

import (
	"context"
	"errors"
	"time"

	"sumni-finance-backend/internal/common"
	http2 "sumni-finance-backend/internal/identity/api/http"
	"sumni-finance-backend/internal/identity/app/models"
)

type AuthenticatorStub struct {
	MockLoginURL       func() (http2.AuthRequest, error)
	MockHandleCallback func(ctx context.Context, input http2.CallbackInput) (models.Session, error)
	MockRotateSession  func(ctx context.Context, oldSession models.Session) (models.Session, error)
	MockVerifySession  func(ctx context.Context, session models.Session) error
	MockLogoutURL      func(idTokenHint, postLogoutRedirectURI string) (string, error)
}

// NewAuthenticatorStub instantiates a baseline implementation that returns successful, safe dummy entries by default.
func NewAuthenticatorStub() *AuthenticatorStub {
	return &AuthenticatorStub{
		MockLoginURL: func() (http2.AuthRequest, error) {
			return http2.AuthRequest{
				Url:      "https://mock-keycloak.local/auth?state=mocked-state",
				State:    "mocked-state",
				Verifier: "mocked-pkce-verifier-string-value-goes-here",
			}, nil
		},
		MockHandleCallback: func(ctx context.Context, input http2.CallbackInput) (models.Session, error) {
			if input.State != input.ExpectedState {
				return models.Session{}, errors.New("state mismatch, possible CSRF")
			}
			return models.Session{
				ID:                 models.SessionID{UUID: common.NewUUIDv7()},
				UserSubject:        "user_stub_99999",
				IDToken:            "mocked.id.token",
				AccessToken:        "mocked.access.token",
				RefreshToken:       "mocked.refresh.token",
				IsExpired:          false,
				ActiveOrganization: "MockOrg",
				ActiveOffice:       "HQ-Main",
				ExpiresAt:          time.Now().Add(1 * time.Hour),
				CreatedAt:          time.Now(),
			}, nil
		},
		MockRotateSession: func(ctx context.Context, oldSession models.Session) (models.Session, error) {
			if oldSession.RefreshToken == "" {
				return models.Session{}, models.ErrMissingRefreshToken
			}
			// Simulate a dead refresh token scenario if explicitly flagged in test contexts
			if oldSession.RefreshToken == "expired.refresh.token" {
				return models.Session{}, models.ErrTokenInvalidated
			}
			if oldSession.RefreshToken == "network.fault.token" {
				return models.Session{}, models.ErrVerificationFailed
			}

			return models.Session{
				ID:                 models.SessionID{UUID: common.NewUUIDv7()},
				UserSubject:        oldSession.UserSubject,
				IDToken:            "new.mocked.id.token",
				AccessToken:        "new.mocked.access.token",
				RefreshToken:       "new.mocked.refresh.token",
				IsExpired:          false,
				ActiveOrganization: oldSession.ActiveOrganization,
				ActiveOffice:       oldSession.ActiveOffice,
				ExpiresAt:          time.Now().Add(1 * time.Hour),
				UpdatedAt:          common.ToPtr(time.Now()),
			}, nil
		},
		MockVerifySession: func(ctx context.Context, session models.Session) error {
			if time.Now().Add(15 * time.Second).After(session.ExpiresAt) {
				return errors.New("token expired")
			}
			if session.AccessToken == "invalid.access.token" {
				return models.ErrTokenInvalidated
			}
			if session.AccessToken == "network.fault.token" {
				return models.ErrVerificationFailed
			}
			return nil
		},
		MockLogoutURL: func(idTokenHint, postLogoutRedirectURI string) (string, error) {
			return "https://mock-keycloak.local/realms/mock/protocol/openid-connect/logout?post_logout_redirect_uri=" + postLogoutRedirectURI, nil
		},
	}
}

func (s *AuthenticatorStub) LoginURL() (http2.AuthRequest, error) {
	return s.MockLoginURL()
}

func (s *AuthenticatorStub) HandleCallback(ctx context.Context, input http2.CallbackInput) (models.Session, error) {
	return s.MockHandleCallback(ctx, input)
}

func (s *AuthenticatorStub) RotateSession(ctx context.Context, oldSession models.Session) (models.Session, error) {
	return s.MockRotateSession(ctx, oldSession)
}

func (s *AuthenticatorStub) VerifySession(ctx context.Context, session models.Session) error {
	return s.MockVerifySession(ctx, session)
}

func (s *AuthenticatorStub) LogoutURL(idTokenHint, postLogoutRedirectURI string) (string, error) {
	return s.MockLogoutURL(idTokenHint, postLogoutRedirectURI)
}
