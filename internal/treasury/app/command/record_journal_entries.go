package command

import (
	"context"
	"fmt"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/shopspring/decimal"
)

type RecordedJournalEntriesCmd struct {
	FundSourceUUID    domain.FundSourceUUID
	JournalEntryItems []JournalEntryItem
}

type JournalEntryItem struct {
	Amount      decimal.Decimal
	Description *string

	EntryType shared.EntryType

	TransactionDate time.Time
	TransactionNo   *string
}

func (h *Handlers) RecordJournalEntries(ctx context.Context, cmd RecordedJournalEntriesCmd) error {
	if len(cmd.JournalEntryItems) == 0 {
		return common.NewInvalidInputError("empty-journal-items", "journal items can't be empty")
	}

	return h.fundSourceRepository.RecordJournalEntries(
		ctx,
		cmd.FundSourceUUID,
		func(fundSource *domain.FundSource) ([]*domain.JournalEntry, error) {
			journalEntries := make([]*domain.JournalEntry, 0, len(cmd.JournalEntryItems))
			for _, item := range cmd.JournalEntryItems {
				snapshot, err := recordJournalEntryToFundSource(
					fundSource,
					item.Amount,
					item.EntryType,
				)
				if err != nil {
					return nil, common.NewInvalidInputError("invalid-balance-operation", "%s", err.Error())
				}

				journalEntry, err := domain.NewJournalEntry(
					item.Amount,
					item.EntryType,
					item.TransactionDate,
					item.TransactionNo,
					item.Description,
					cmd.FundSourceUUID,
					snapshot.before,
				)
				if err != nil {
					return nil, err
				}

				err = journalEntry.SetBalanceAfter(snapshot.after)
				if err != nil {
					return nil, err
				}

				journalEntries = append(journalEntries, journalEntry)
			}

			return journalEntries, nil
		},
	)
}

type balanceSnapshot struct {
	before decimal.Decimal
	after  decimal.Decimal
}

func recordJournalEntryToFundSource(
	fs *domain.FundSource,
	amount decimal.Decimal,
	entryType shared.EntryType,
) (balanceSnapshot, error) {
	money, err := shared.NewMoney(amount, fs.Currency())
	if err != nil {
		return balanceSnapshot{}, err
	}

	snapshot := balanceSnapshot{
		before: fs.Balance().Amount(),
	}

	switch entryType {
	case shared.EntryTypeDebit:
		err = fs.Withdraw(money)
	case shared.EntryTypeCredit:
		err = fs.TopUp(money)
	default:
		return balanceSnapshot{}, fmt.Errorf("unsupported entry type: %s", entryType)
	}
	if err != nil {
		return balanceSnapshot{}, err
	}

	snapshot.after = fs.Balance().Amount()
	return snapshot, nil
}
