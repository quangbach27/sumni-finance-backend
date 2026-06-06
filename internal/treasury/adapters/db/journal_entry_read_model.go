package db

import (
	"context"
	"fmt"

	"sumni-finance-backend/internal/treasury/adapters/db/dbmodels"
	"sumni-finance-backend/internal/treasury/app/query"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JournalEntryReadModel struct {
	db *pgxpool.Pool
}

func NewJournalEntryReadModel(db *pgxpool.Pool) *JournalEntryReadModel {
	if db == nil {
		panic("db can't be nil")
	}

	return &JournalEntryReadModel{db: db}
}

func (r *JournalEntryReadModel) ListJournalEntries(
	ctx context.Context,
	q query.ListJournalEntriesQuery,
) (query.PaginatedResult[query.JournalEntryView], error) {
	queries := dbmodels.New(r.db)

	totalItems, err := queries.CountJournalEntries(ctx, dbmodels.CountJournalEntriesParams{
		FundSourceUuid: q.FundSourceUUID,
		DateFrom:       q.DateFrom,
		DateTo:         q.DateTo,
	})
	if err != nil {
		return query.PaginatedResult[query.JournalEntryView]{}, fmt.Errorf("error counting journal entries: %w", err)
	}

	rows, err := queries.ListJournalEntries(ctx, dbmodels.ListJournalEntriesParams{
		FundSourceUuid: q.FundSourceUUID,
		DateFrom:       q.DateFrom,
		DateTo:         q.DateTo,
		PageLimit:      int32(q.PageSize),
		PageOffset:     int32((q.Page - 1) * q.PageSize),
	})
	if err != nil {
		return query.PaginatedResult[query.JournalEntryView]{}, fmt.Errorf("error listing journal entries: %w", err)
	}

	views := make([]query.JournalEntryView, 0, len(rows))
	for _, row := range rows {
		views = append(views, toJournalEntryView(row))
	}

	return query.PaginatedResult[query.JournalEntryView]{
		Items:      views,
		TotalItems: int(totalItems),
		Page:       q.Page,
		PageSize:   q.PageSize,
	}, nil
}

func toJournalEntryView(model dbmodels.TreasuryJournalEntry) query.JournalEntryView {
	return query.JournalEntryView{
		UUID:             model.JournalEntryUuid,
		Amount:           model.Amount,
		EntryType:        model.EntryType,
		TransactionDate:  model.TransactionDate,
		TransactionNo:    model.TransactionNo,
		Description:      model.Description,
		Status:           model.Status,
		FundSourceUUID:   model.FundSourceUuid,
		BalanceBefore:    model.BalanceBefore,
		BalanceAfter:     model.BalanceAfter,
		VoidedBy:         model.VoidedBy,
		VoidedAt:         model.VoidedAt,
		VoidedReason:     model.VoidedReason,
		ReverseEntryUUID: model.ReverseEntryUuid,
		CreatedAt:        model.CreatedAt,
		CreatedBy:        model.CreatedBy,
		UpdatedAt:        model.UpdatedAt,
		UpdatedBy:        model.UpdatedBy,
	}
}
