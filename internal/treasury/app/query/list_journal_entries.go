package query

import (
	"context"

	"sumni-finance-backend/internal/common"
)

func (h *Handlers) ListJournalEntries(
	ctx context.Context,
	query ListJournalEntriesQuery,
) (PaginatedResult[JournalEntryView], error) {
	if query.FundSourceUUID.IsZero() {
		return PaginatedResult[JournalEntryView]{}, common.NewInvalidInputError("empty-fund-provider-uuid", "fund provider uuid can't be empty")
	}

	return h.journalEntryReadModel.ListJournalEntries(ctx, query)
}
