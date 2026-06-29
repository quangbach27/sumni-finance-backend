package keycloak

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"sumni-finance-backend/internal/common"
	http2 "sumni-finance-backend/internal/identity/api/http"
	"sumni-finance-backend/internal/identity/app/models"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	// tokenRefreshBuffer provides a safety window to proactively refresh tokens
	// before they physically expire, preventing mid-flight network request failures.
	tokenRefreshBuffer = 15 * time.Second
)

var defaultScope = []string{oidc.ScopeOpenID, "profile", "email", "roles", "organization"}

type Authenticator struct {
	provider *oidc.Provider
	oauth2   oauth2.Config
	config   AuthenticatorConfig
}

type AuthenticatorConfig struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	HttpClient   *http.Client
}

func (c AuthenticatorConfig) issueURL() string {
	return fmt.Sprintf("%s/realms/%s", strings.TrimRight(c.BaseURL, "/"), c.Realm)
}

func NewAuthenticator(ctx context.Context, config AuthenticatorConfig) (*Authenticator, error) {
	oidcCtx := ctx
	if config.HttpClient != nil {
		oidcCtx = context.WithValue(oidcCtx, oauth2.HTTPClient, config.HttpClient)
	}

	provider, err := oidc.NewProvider(oidcCtx, config.issueURL())
	if err != nil {
		return nil, fmt.Errorf("oidc discovery failed: %w", err)
	}

	scopes := config.Scopes
	if len(scopes) == 0 {
		scopes = defaultScope
	}

	return &Authenticator{
		config:   config,
		provider: provider,
		oauth2: oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       scopes,
		},
	}, nil
}

func (a *Authenticator) LoginURL() (http2.AuthRequest, error) {
	state, err := randomString(32)
	if err != nil {
		return http2.AuthRequest{}, fmt.Errorf("failed to generate state: %w", err)
	}

	verifier := oauth2.GenerateVerifier()
	url := a.oauth2.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	return http2.AuthRequest{
		Url:      url,
		State:    state,
		Verifier: verifier,
	}, nil
}

func (a *Authenticator) HandleCallback(ctx context.Context, input http2.CallbackInput) (models.Session, error) {
	if input.Error != "" {
		return models.Session{}, fmt.Errorf("authorization failed: %s: %s", input.Error, input.ErrorDescription)
	}
	if input.Code == "" {
		return models.Session{}, errors.New("missing code parameter")
	}
	if input.State == "" || input.ExpectedState == "" {
		return models.Session{}, errors.New("missing state")
	}
	if input.State != input.ExpectedState {
		return models.Session{}, errors.New("state mismatch, possible CSRF")
	}
	if input.Verifier == "" {
		return models.Session{}, errors.New("missing pkce verifier")
	}

	token, err := a.oauth2.Exchange(
		a.contextWithClient(ctx),
		input.Code,
		oauth2.VerifierOption(input.Verifier),
	)
	if err != nil {
		return models.Session{}, fmt.Errorf("token exchnage failed: %w", err)
	}

	rawIDToken, err := a.verifyIdToken(ctx, token)
	if err != nil {
		return models.Session{}, err
	}

	verifiedAccessToken, err := a.verifyAccessToken(ctx, token.AccessToken)
	if err != nil {
		return models.Session{}, err
	}

	tokenClaims, err := models.UnmarshalTokenClaims(verifiedAccessToken)
	if err != nil {
		return models.Session{}, err
	}

	return models.Session{
		ID:                 models.SessionID{UUID: common.NewUUIDv7()},
		UserSubject:        tokenClaims.Subject(),
		IDToken:            rawIDToken,
		AccessToken:        token.AccessToken,
		RefreshToken:       token.RefreshToken,
		IsExpired:          false,
		ActiveOrganization: tokenClaims.Organization().Name(),
		ActiveOffice:       tokenClaims.Organization().Groups()[0],
		ExpiresAt:          token.Expiry,
		CreatedAt:          time.Now(),
	}, nil
}

func (a *Authenticator) RotateSession(ctx context.Context, oldSession models.Session) (models.Session, error) {
	if oldSession.RefreshToken == "" {
		return models.Session{}, models.ErrMissingRefreshToken
	}

	ctx = a.contextWithClient(ctx)

	src := a.oauth2.TokenSource(ctx, &oauth2.Token{RefreshToken: oldSession.RefreshToken})
	token, err := src.Token()
	if err != nil {
		var retrieveErr *oauth2.RetrieveError
		if errors.As(err, &retrieveErr) {
			if retrieveErr.Response.StatusCode == http.StatusBadRequest {
				return models.Session{}, models.ErrTokenInvalidated
			}
		}
		return models.Session{}, err
	}

	rawIDToken, err := a.verifyIdToken(ctx, token)
	if err != nil {
		return models.Session{}, err
	}

	_, err = a.verifyAccessToken(ctx, token.AccessToken)
	if err != nil {
		return models.Session{}, err
	}

	return models.Session{
		ID:                 models.SessionID{UUID: common.NewUUIDv7()}, // regenerate sessionID
		UserSubject:        oldSession.UserSubject,
		IDToken:            rawIDToken,
		AccessToken:        token.AccessToken,
		RefreshToken:       token.RefreshToken,
		IsExpired:          false,
		ActiveOrganization: oldSession.ActiveOrganization,
		ActiveOffice:       oldSession.ActiveOffice,
		ExpiresAt:          token.Expiry,
		UpdatedAt:          common.ToPtr(time.Now()),
	}, nil
}

func (a *Authenticator) VerifySession(
	ctx context.Context,
	session models.Session,
) error {
	if time.Now().Add(tokenRefreshBuffer).After(session.ExpiresAt) {
		return errors.New("token expired")
	}

	if _, err := a.verifyAccessToken(ctx, session.AccessToken); err != nil {
		return fmt.Errorf("accessToken verification failed: %w", err)
	}

	return nil
}

func (a *Authenticator) verifyAccessToken(
	ctx context.Context,
	rawAccessToken string,
) (*oidc.IDToken, error) {
	return a.provider.
		Verifier(&oidc.Config{SkipClientIDCheck: true}).
		Verify(ctx, rawAccessToken)
}

func (a *Authenticator) LogoutURL(idTokenHint, postLogoutRedirectURI string) (string, error) {
	var claims struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := a.provider.Claims(&claims); err != nil {
		return "", fmt.Errorf("failed to read end_session_endpoint: %w", err)
	}
	if claims.EndSessionEndpoint == "" {
		return "", errors.New("provider does not advertise an end_session_endpoint")
	}

	return buildLogoutURL(claims.EndSessionEndpoint, idTokenHint, postLogoutRedirectURI)
}

func buildLogoutURL(endSessionEndpoint, idTokenHint, postLogoutRedirectURI string) (string, error) {
	u, err := url.Parse(endSessionEndpoint)
	if err != nil {
		return "", err
	}

	q := u.Query()
	if idTokenHint != "" {
		q.Set("id_token_hint", idTokenHint)
	}
	if postLogoutRedirectURI != "" {
		q.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (a Authenticator) verifyIdToken(ctx context.Context, token *oauth2.Token) (string, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return "", errors.New("no id_token in token response")
	}

	_, err := a.provider.
		Verifier(&oidc.Config{ClientID: a.oauth2.ClientID}).
		Verify(ctx, rawIDToken)
	if err != nil {
		return "", fmt.Errorf("id_token handshake verification failed: %w", err)
	}

	return rawIDToken, nil
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (a *Authenticator) contextWithClient(ctx context.Context) context.Context {
	if a.config.HttpClient != nil {
		return context.WithValue(ctx, oauth2.HTTPClient, a.config.HttpClient)
	}

	return ctx
}
