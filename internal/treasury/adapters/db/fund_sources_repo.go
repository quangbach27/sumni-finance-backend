package db

import (
	"context"
	"errors"
	"fmt"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/adapters/db/dbmodels"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FundSourceRepo struct {
	db *pgxpool.Pool
}

func NewFundSourceRepository(db *pgxpool.Pool) *FundSourceRepo {
	if db == nil {
		panic("db can't be nil")
	}

	return &FundSourceRepo{db: db}
}

func (r *FundSourceRepo) getFundSourceByUUID(
	ctx context.Context,
	queries *dbmodels.Queries,
	fundSourceUUID domain.FundSourceUUID,
) (*domain.FundSource, error) {
	fsModel, err := queries.GetFundSourceByUUID(ctx, fundSourceUUID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving fund source: %w", err)
	}

	return unmarshalFundSourceFromDb(fsModel), nil
}

func unmarshalFundSourceFromDb(fsModel dbmodels.TreasuryFundSource) *domain.FundSource {
	balance := shared.UnmarshalMoney(fsModel.Balance, fsModel.Currency)
	availableBalance := shared.UnmarshalMoney(fsModel.AvailableBalance, fsModel.Currency)
	audit := shared.UnmarshalAudit(fsModel.CreatedAt, fsModel.CreatedBy, fsModel.UpdatedAt, fsModel.UpdatedBy)

	var metadata domain.FundSourceMetadata
	switch fsModel.SourceType {
	case domain.FundSourceTypeBank:
		metadata = domain.UnmarshalFundSourceBankMetadata(
			fsModel.BankInfo,
			common.SafeDeref(fsModel.BankAccountNumber, ""),
			common.SafeDeref(fsModel.BankAccountOwner, ""),
		)
	case domain.FundSourceTypeCash:
		metadata = domain.UnmarshalFundSourceCashMetadata(
			common.SafeDeref(fsModel.CashOwner, ""),
		)
	}

	return domain.UnmarshalFundSource(
		fsModel.FundSourceUuid,
		fsModel.Name,
		fsModel.SourceType,
		balance,
		availableBalance,
		fsModel.Currency,
		metadata,
		audit,
	)
}

func (r *FundSourceRepo) SaveFundSource(ctx context.Context, fundSource *domain.FundSource) error {
	if fundSource == nil {
		return errors.New("fund source can't be empty")
	}

	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		params := dbmodels.InsertFundSourceParams{
			FundSourceUuid:   fundSource.UUID(),
			Name:             fundSource.Name(),
			SourceType:       fundSource.SourceType(),
			Balance:          fundSource.Balance().Amount(),
			AvailableBalance: fundSource.AvailableBalance().Amount(),
			Currency:         fundSource.Currency(),
			CreatedAt:        fundSource.CreatedAt(),
			CreatedBy:        fundSource.CreatedBy(),
		}
		addFundSourceMetadataToParams(fundSource, &params)

		err := queries.InsertFundSource(ctx, params)
		if err != nil {
			return fmt.Errorf("error saving fund source: %w", err)
		}

		return nil
	})
}

func addFundSourceMetadataToParams(fs *domain.FundSource, params *dbmodels.InsertFundSourceParams) {
	if bankMetadata, ok := fs.BankMetadata(); ok {
		params.BankInfo = bankMetadata.BankInfo()
		params.BankCode = common.ToPtr(bankMetadata.BankInfo().BankCode())
		params.BankAccountNumber = common.ToPtr(bankMetadata.AccountNumber())
		params.BankAccountOwner = common.ToPtr(bankMetadata.AccountOwner())
	}

	if cashMetadata, ok := fs.CashMetadata(); ok {
		params.CashOwner = common.ToPtr(cashMetadata.OwnerName())
	}
}

func (r *FundSourceRepo) RecordJournalEntries(
	ctx context.Context,
	fundSourceUUID domain.FundSourceUUID,
	recordFn func(fundSource *domain.FundSource) ([]*domain.JournalEntry, error),
) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		fundSource, err := r.getFundSourceByUUID(ctx, queries, fundSourceUUID)
		if err != nil {
			return err
		}

		journalEntries, err := recordFn(fundSource)
		if err != nil {
			return err
		}

		if err = r.updateFundSource(ctx, queries, fundSource); err != nil {
			return err
		}

		return r.insertJournalEntries(ctx, queries, journalEntries)
	})
}

func (r *FundSourceRepo) updateFundSource(
	ctx context.Context,
	queries *dbmodels.Queries,
	fs *domain.FundSource,
) error {
	params := dbmodels.UpdateFundSourceParams{
		FundSourceUuid:   fs.UUID(),
		Balance:          fs.Balance().Amount(),
		AvailableBalance: fs.AvailableBalance().Amount(),
		UpdatedAt:        fs.UpdatedAt(),
		UpdatedBy:        fs.UpdatedBy(),
	}
	if err := queries.UpdateFundSource(ctx, params); err != nil {
		return fmt.Errorf("error updating fund source: %w", err)
	}
	return nil
}

func (r *FundSourceRepo) updateJournalEntry(
	ctx context.Context,
	queries *dbmodels.Queries,
	entry *domain.JournalEntry,
) error {
	params := dbmodels.UpdateJournalEntryParams{
		JournalEntryUuid: entry.UUID(),
		Status:           entry.Status(),
		ReverseEntryUuid: entry.ReverseEntryUUID(),
		VoidedBy:         entry.VoidedBy(),
		VoidedAt:         entry.VoidedAt(),
		VoidedReason:     entry.VoidedReason(),
		UpdatedAt:        entry.UpdatedAt(),
		UpdatedBy:        entry.UpdatedBy(),
	}
	if err := queries.UpdateJournalEntry(ctx, params); err != nil {
		return fmt.Errorf("error updating journal entry: %w", err)
	}
	return nil
}

func (r *FundSourceRepo) insertJournalEntries(
	ctx context.Context,
	queries *dbmodels.Queries,
	entries []*domain.JournalEntry,
) error {
	params, err := buildBatchInsertJournalEntriesParams(entries)
	if err != nil {
		return err
	}
	if _, err = queries.BatchInsertJournalEntries(ctx, params); err != nil {
		return fmt.Errorf("error inserting journal entries: %w", err)
	}
	return nil
}

func buildBatchInsertJournalEntriesParams(entries []*domain.JournalEntry) ([]dbmodels.BatchInsertJournalEntriesParams, error) {
	if len(entries) == 0 {
		return nil, errors.New("journal entries can't be empty")
	}
	params := make([]dbmodels.BatchInsertJournalEntriesParams, len(entries))
	for i, e := range entries {
		if e == nil {
			return nil, fmt.Errorf("journal entry at index %d is nil", i)
		}
		params[i] = dbmodels.BatchInsertJournalEntriesParams{
			JournalEntryUuid: e.UUID(),
			FundSourceUuid:   e.FundSourceUUID(),
			Amount:           e.Amount(),
			EntryType:        e.EntryType(),
			BalanceBefore:    e.BalanceBefore(),
			BalanceAfter:     e.BalanceAfter(),
			Status:           e.Status(),
			TransactionDate:  e.TransactionDate(),
			TransactionNo:    e.TransactionNo(),
			Description:      e.Description(),
			CreatedAt:        e.CreatedAt(),
			CreatedBy:        e.CreatedBy(),
		}
	}
	return params, nil
}

func (r *FundSourceRepo) VoidJournalEntry(
	ctx context.Context,
	fundSourceUUID domain.FundSourceUUID,
	journalEntryUUIDToVoid domain.JournalEntryUUID,
	voidFn func(
		fundSource *domain.FundSource,
		journalEntryToVoid *domain.JournalEntry,
	) (*domain.JournalEntry, error),
) (domain.JournalEntryUUID, error) {
	var reverseJournalEntryUUID domain.JournalEntryUUID

	err := common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		fundSource, err := r.getFundSourceByUUID(ctx, queries, fundSourceUUID)
		if err != nil {
			return err
		}

		journalEntryToVoid, err := r.getJournalEntryByUUID(ctx, queries, journalEntryUUIDToVoid)
		if err != nil {
			return err
		}

		reverseEntry, err := voidFn(fundSource, journalEntryToVoid)
		if err != nil {
			return fmt.Errorf("fail to void fn: %w", err)
		}

		reverseJournalEntryUUID = reverseEntry.UUID()

		if err = r.updateFundSource(ctx, queries, fundSource); err != nil {
			return err
		}

		if err = r.updateJournalEntry(ctx, queries, journalEntryToVoid); err != nil {
			return err
		}

		return r.insertJournalEntries(ctx, queries, []*domain.JournalEntry{reverseEntry})
	})
	if err != nil {
		return domain.JournalEntryUUID{}, err
	}
	
	return reverseJournalEntryUUID, nil
}

func (r *FundSourceRepo) getJournalEntryByUUID(
	ctx context.Context,
	queries *dbmodels.Queries,
	journalEntryUUID domain.JournalEntryUUID,
) (*domain.JournalEntry, error) {
	entryModel, err := queries.GetJournalEntryByUID(ctx, journalEntryUUID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving journal entry by UUID: %w", err)
	}

	return unmarshalJournalEntryFromDb(entryModel), nil
}

func unmarshalJournalEntryFromDb(model dbmodels.TreasuryJournalEntry) *domain.JournalEntry {
	audit := shared.UnmarshalAudit(model.CreatedAt, model.CreatedBy, model.UpdatedAt, model.UpdatedBy)

	return domain.UnmarshalJournalEntry(
		model.JournalEntryUuid,
		model.Amount,
		model.EntryType,
		model.TransactionDate,
		model.TransactionNo,
		model.Description,
		model.Status,
		model.FundSourceUuid,
		model.BalanceBefore,
		model.BalanceAfter,
		model.VoidedBy,
		model.VoidedAt,
		model.VoidedReason,
		model.ReverseEntryUuid,
		audit,
	)
}
