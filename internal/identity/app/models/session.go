package models

import (
	"context"
	"errors"
	"time"

	"sumni-finance-backend/internal/common"
)

type SessionStore interface {
	GetActiveSessionByID(ctx context.Context, sessionID SessionID) (Session, error)
	UpsertSession(ctx context.Context, session Session) error
	ExpireSessionByID(ctx context.Context, sessionID SessionID) error
}

var ErrActiveSessionNotFound = errors.New("active session not found")

type SessionManager interface {
	VerifySession(ctx context.Context, session Session) error
	RotateSession(ctx context.Context, oldSession Session) (Session, error)
}

var (
	ErrTokenInvalidated    = errors.New("refresh token is expired, replayed, or revoked")
	ErrMissingRefreshToken = errors.New("missing refresh token")
	ErrVerificationFailed  = errors.New("crypto token verification failed due to network or configuration issue")
)

type SessionID struct {
	common.UUID
}

type Session struct {
	ID          SessionID
	UserSubject string

	IDToken      string
	AccessToken  string
	RefreshToken string

	ActiveOrganization string
	ActiveOffice       string

	IsExpired bool

	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt *time.Time
}
