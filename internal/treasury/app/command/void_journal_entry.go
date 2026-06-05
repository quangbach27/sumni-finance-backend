package command

import (
	"context"
	"time"

	"sumni-finance-backend/internal/treasury/domain"
)

type VoidJournalEntryCmd struct {
	FundSourceUUID         domain.FundSourceUUID
	JournalEntryUUIDToVoid domain.JournalEntryUUID
	VoidedBy               string
	VoidedReason           string
}

func (h *Handler) VoidJournalEntry(
	ctx context.Context,
	cmd VoidJournalEntryCmd,
) (*domain.JournalEntry, error) {
	return h.fundSourceRepository.VoidJournalEntry(
		ctx,
		cmd.FundSourceUUID,
		cmd.JournalEntryUUIDToVoid,
		func(fundSource *domain.FundSource, je *domain.JournalEntry) (*domain.JournalEntry, error) {
			reverse, err := je.Void(cmd.VoidedBy, time.Now(), cmd.VoidedReason)
			if err != nil {
				return nil, err
			}

			balanceSnapshot, err := recordJournalEntryToFundSource(fundSource, reverse.Amount(), reverse.EntryType())
			if err != nil {
				return nil, err
			}

			reverse.SetBalanceBefore(balanceSnapshot.before)
			reverse.SetBalanceAfter(balanceSnapshot.after)

			return reverse, nil
		},
	)
}
