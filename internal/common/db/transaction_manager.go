package db

import (
	"context"
	"errors"
	"fmt"
	"sumni-finance-backend/internal/common/logs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxTransactionManager struct {
	pgxPool *pgxpool.Pool
}

func NewPgxTransactionManager(pgxPool *pgxpool.Pool) *PgxTransactionManager {
	return &PgxTransactionManager{
		pgxPool: pgxPool,
	}
}

func (tm *PgxTransactionManager) WithTx(
	ctx context.Context,
	fn func(tx pgx.Tx) error,
) (err error) {
	var tx pgx.Tx
	tx, err = tm.pgxPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		err = tm.finishTransaction(ctx, tx, err)
	}()

	return fn(tx)
}

func (tm *PgxTransactionManager) finishTransaction(ctx context.Context, tx pgx.Tx, err error) error {
	logger := logs.FromContext(ctx)

	if err == nil {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return fmt.Errorf("commit transaction failed after successful operation: %w", commitErr)
		}

		logger.Info("commit transaction successful")
		return nil
	}

	if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
		return errors.Join(fmt.Errorf("rollback transaction failed: %w", rollbackErr), err)
	}

	logger.Info("rollback transaction successful")
	return err
}
