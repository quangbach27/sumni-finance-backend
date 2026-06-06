package query

import (
	"context"
	"time"

	"sumni-finance-backend/internal/treasury/domain"
)

type FundSourceReadModel interface {
	ListFundSources(ctx context.Context) ([]FundSourceView, error)
}

type ListJournalEntriesQuery struct {
	Page           int
	PageSize       int
	FundSourceUUID domain.FundSourceUUID
	DateFrom       *time.Time
	DateTo         *time.Time
}

type JournalEntryReadModel interface {
	ListJournalEntries(ctx context.Context, query ListJournalEntriesQuery) (PaginatedResult[JournalEntryView], error)
}

type Handlers struct {
	fundSourceReadModel   FundSourceReadModel
	journalEntryReadModel JournalEntryReadModel
}

func NewHandlers(
	fundSourceReadModel FundSourceReadModel,
	journalEntryReadModel JournalEntryReadModel,
) *Handlers {
	if fundSourceReadModel == nil {
		panic("fund source read model can't be nil")
	}

	if journalEntryReadModel == nil {
		panic("journal source read model can't be nil")
	}

	return &Handlers{
		fundSourceReadModel:   fundSourceReadModel,
		journalEntryReadModel: journalEntryReadModel,
	}
}
