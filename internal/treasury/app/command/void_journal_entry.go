package command

import (
	"context"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/treasury/domain"
)

type VoidJournalEntryCmd struct {
	FundSourceUUID         domain.FundSourceUUID
	JournalEntryUUIDToVoid domain.JournalEntryUUID
	VoidedBy               string
	VoidedReason           string
}

func (h *Handlers) VoidJournalEntry(
	ctx context.Context,
	cmd VoidJournalEntryCmd,
) (*domain.JournalEntry, error) {
	return h.fundSourceRepository.VoidJournalEntry(
		ctx,
		cmd.FundSourceUUID,
		cmd.JournalEntryUUIDToVoid,
		func(fundSource *domain.FundSource, je *domain.JournalEntry) (*domain.JournalEntry, error) {
			if !fundSource.UUID().Equals(je.FundSourceUUID().UUID) {
				return nil, common.NewInvalidInputError(
					"fund-source-mismatch",
					"journal entry %s does not belong to fund source %s",
					je.FundSourceUUID().UUID, fundSource.UUID(),
				)
			}

			journalEntryReverse, err := je.Void(cmd.VoidedBy, time.Now(), cmd.VoidedReason)
			if err != nil {
				return nil, err
			}

			balanceSnapshot, err := recordJournalEntryToFundSource(
				fundSource,
				journalEntryReverse.Amount(),
				journalEntryReverse.EntryType(),
			)
			if err != nil {
				return nil, err
			}

			journalEntryReverse.SetBalanceBefore(balanceSnapshot.before)
			err = journalEntryReverse.SetBalanceAfter(balanceSnapshot.after)
			if err != nil {
				return nil, err
			}

			return journalEntryReverse, nil
		},
	)
}
