package db

import (
	"context"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/finance/app/models"
	"sumni-finance-backend/internal/finance/db/dbmodels"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OfficeRepo struct {
	db *pgxpool.Pool
}

func NewOfficeRepo(db *pgxpool.Pool) *OfficeRepo {
	if db == nil {
		panic("db can't be empty")
	}

	return &OfficeRepo{db: db}
}

func (r *OfficeRepo) OfficeByUUID(ctx context.Context, uuid models.OfficeUUID) (models.Office, error) {
	queries := dbmodels.New(r.db)

	model, err := queries.OfficeByUUID(ctx, uuid)
	if err != nil {
		return models.Office{}, fmt.Errorf("failed to get office uuid: %w", err)
	}

	return models.Office{
		UUID: model.OfficeUuid,
		Name: model.Name,
	}, nil
}

func (r *OfficeRepo) SaveOffice(ctx context.Context, office models.Office) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		err := queries.UpsertOffice(ctx, dbmodels.UpsertOfficeParams{
			OfficeUuid: office.UUID,
			Name:       office.Name,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert office: %w", err)
		}

		return nil
	})
}
