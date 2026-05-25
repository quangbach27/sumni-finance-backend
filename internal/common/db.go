package common

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/jackc/pgx/v5"
)

const couldNotSerializeAccessErrMsg = "could not serialize access"

type TxBeginer interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

func UpdateInReadCommitTx(
	ctx context.Context,
	db TxBeginer,
	fn func(ctx context.Context, tx pgx.Tx) error,
) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = time.Millisecond
	b.MaxInterval = 500 * time.Millisecond
	b.Multiplier = 2.0
	b.RandomizationFactor = 0.5

	_, err := backoff.Retry(
		ctx,
		func() (struct{}, error) {
			err := updateInTx(ctx, db, pgx.ReadCommitted, fn)
			if err != nil {
				if strings.Contains(err.Error(), couldNotSerializeAccessErrMsg) {
					// Retryable
					return struct{}{}, err
				} else {
					return struct{}{}, backoff.Permanent(err)
				}
			}

			return struct{}{}, nil
		},
		backoff.WithBackOff(b),
		backoff.WithMaxElapsedTime(5),
	)

	return err
}

func UpdateInTx(
	ctx context.Context,
	db TxBeginer,
	fn func(ctx context.Context, tx pgx.Tx) error,
) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = time.Millisecond
	b.MaxInterval = 500 * time.Millisecond
	b.Multiplier = 2.0
	b.RandomizationFactor = 0.5

	_, err := backoff.Retry(
		ctx,
		func() (struct{}, error) {
			err := updateInTx(ctx, db, pgx.RepeatableRead, fn)
			if err != nil {
				if strings.Contains(err.Error(), couldNotSerializeAccessErrMsg) {
					// Retryable
					return struct{}{}, err
				} else {
					return struct{}{}, backoff.Permanent(err)
				}
			}

			return struct{}{}, nil
		},
		backoff.WithBackOff(b),
		backoff.WithMaxElapsedTime(5),
	)

	return err
}

func updateInTx(
	ctx context.Context,
	db TxBeginer,
	isoLevel pgx.TxIsoLevel,
	fn func(ctx context.Context, tx pgx.Tx) error,
) (err error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: isoLevel})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("could not begin transaction (possible pool exhaustion — context deadline exceeded): %w", err)
		}
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = errors.Join(err, rollbackErr)
			}
			return
		}

		err = tx.Commit(ctx)
	}()

	return fn(ctx, tx)
}
