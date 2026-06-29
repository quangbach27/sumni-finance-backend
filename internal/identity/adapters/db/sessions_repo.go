package db

import (
	"context"
	"errors"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/identity/adapters/db/dbmodels"
	"sumni-finance-backend/internal/identity/app/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepo struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepo {
	if db == nil {
		panic("db can't be empty")
	}

	return &SessionRepo{
		db: db,
	}
}

func (r *SessionRepo) GetActiveSessionByID(ctx context.Context, sessionID models.SessionID) (models.Session, error) {
	queries := dbmodels.New(r.db)

	sessionDb, err := queries.GetActiveSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Session{}, models.ErrActiveSessionNotFound
		}

		return models.Session{}, fmt.Errorf("failed to get session: %w", err)
	}

	return unmarshalSessionToModel(sessionDb), nil
}

func (r *SessionRepo) UpsertSession(ctx context.Context, session models.Session) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		err := queries.UpsertSession(ctx, dbmodels.UpsertSessionParams{
			SessionID:          session.ID,
			UserSubject:        session.UserSubject,
			IDToken:            session.IDToken,
			AccessToken:        session.AccessToken,
			RefreshToken:       session.RefreshToken,
			ActiveOrganization: session.ActiveOrganization,
			ActiveOffice:       session.ActiveOffice,
			IsExpired:          session.IsExpired,
			ExpiresAt:          session.ExpiresAt,
			CreatedAt:          session.CreatedAt,
			UpdatedAt:          session.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert session: %w", err)
		}

		return nil
	})
}

func (r *SessionRepo) ExpireSessionByID(ctx context.Context, sessionID models.SessionID) error {
	queries := dbmodels.New(r.db)

	err := queries.ExpireSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to expired session: %w", err)
	}

	return nil
}

func unmarshalSessionToModel(sessionDb dbmodels.IdentitiesSession) models.Session {
	return models.Session{
		ID:                 sessionDb.SessionID,
		UserSubject:        sessionDb.UserSubject,
		IDToken:            sessionDb.IDToken,
		AccessToken:        sessionDb.AccessToken,
		RefreshToken:       sessionDb.RefreshToken,
		ActiveOrganization: sessionDb.ActiveOrganization,
		ActiveOffice:       sessionDb.ActiveOffice,
		IsExpired:          sessionDb.IsExpired,
		ExpiresAt:          sessionDb.ExpiresAt,
		CreatedAt:          sessionDb.CreatedAt,
		UpdatedAt:          sessionDb.UpdatedAt,
	}
}
