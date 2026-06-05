package http

import (
	"context"

	"sumni-finance-backend/internal/treasury/app/command"
	"sumni-finance-backend/internal/treasury/domain"
)

type Handler struct {
	commandHandler *command.Handler
}

func NewHandler(
	commandHandler *command.Handler,
) Handler {
	if commandHandler == nil {
		panic("command handler can't be nil")
	}

	return Handler{
		commandHandler: commandHandler,
	}
}

// Register a new fund source
// (POST /v1/treasury/fund-sources)
func (h Handler) RegisterFundSource(
	ctx context.Context,
	request RegisterFundSourceRequestObject,
) (RegisterFundSourceResponseObject, error) {
	metadata := request.Body.Metadata

	bankMetadataCmd := command.BankMetadataCmd{}
	cashMetadataCmd := command.CashMetadataCmd{}

	if metadata.BankMetadata != nil {
		bankMetadataCmd.BankCode = metadata.BankMetadata.BankCode
		bankMetadataCmd.AccountNumber = metadata.BankMetadata.AccountNumber
		bankMetadataCmd.AccountOwner = metadata.BankMetadata.AccountOwner
	}
	if metadata.CashMetadata != nil {
		cashMetadataCmd.OwnerName = metadata.CashMetadata.OwnerName
	}

	fundSourceUUID, err := h.commandHandler.RegisterFundSource(ctx, command.RegisterFundSourceCmd{
		Name:            request.Body.Name,
		SourceType:      request.Body.SourceType,
		InitBalance:     request.Body.InitBalance,
		Currency:        request.Body.Currency,
		BankMetadata:    bankMetadataCmd,
		CashMetadataCmd: cashMetadataCmd,
	})
	if err != nil {
		return nil, err
	}

	return RegisterFundSource201JSONResponse{FundSourceUuid: fundSourceUUID}, nil
}

// Add New Journal Entries
// (POST /v1/treasury/fund-sources/{fund_source_uuid}/journal-entries)
func (h Handler) RecordJournalEntries(
	ctx context.Context,
	request RecordJournalEntriesRequestObject,
) (RecordJournalEntriesResponseObject, error) {
	journalEntriesReq := *request.Body
	journalEntryItems := make([]command.JournalEntryItem, 0, len(journalEntriesReq))

	for _, journalEntryReq := range journalEntriesReq {
		journalEntryItems = append(journalEntryItems, command.JournalEntryItem{
			Amount:          journalEntryReq.Amount,
			Description:     journalEntryReq.Description,
			EntryType:       journalEntryReq.EntryType,
			TransactionDate: journalEntryReq.TransactionDate,
			TransactionNo:   journalEntryReq.TransactionNo,
		})
	}

	err := h.commandHandler.RecordJournalEntries(ctx, command.RecordedJournalEntriesCmd{
		FundSourceUUID:    request.FundSourceUuid,
		JournalEntryItems: journalEntryItems,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Void a journal entry by creating a reverse journal entry
// (POST /v1/treasury/fund-sources/{fund_source_uuid}/journal-entries/{journal_entry_uuid}/void)
func (h Handler) VoidJournalEntry(ctx context.Context, request VoidJournalEntryRequestObject) (VoidJournalEntryResponseObject, error) {
	reverseJournalEntry, err := h.commandHandler.VoidJournalEntry(ctx, command.VoidJournalEntryCmd{
		FundSourceUUID:         request.FundSourceUuid,
		JournalEntryUUIDToVoid: request.JournalEntryUuid,
		VoidedReason:           request.Body.VoidReason,
		VoidedBy:               "system",
	})
	if err != nil {
		return nil, err
	}

	return VoidJournalEntry200JSONResponse{
		ReverseJournalEntry: domainJournalEntryToResponse(reverseJournalEntry),
	}, nil
}

func domainJournalEntryToResponse(je *domain.JournalEntry) JournalEntryResponse {
	return JournalEntryResponse{
		JournalEntryUuid: je.UUID(),
		FundSourceUuid:   je.FundSourceUUID(),
		Amount:           je.Amount(),
		EntryType:        je.EntryType(),
		TransactionDate:  je.TransactionDate(),
		TransactionNo:    je.TransactionNo(),
		Description:      je.Description(),
		Status:           je.Status(),
		BalanceBefore:    je.BalanceBefore(),
		BalanceAfter:     je.BalanceAfter(),
		ReverseEntryUuid: je.ReverseEntryUUID(),
		VoidedBy:         je.VoidedBy(),
		VoidedAt:         je.VoidedAt(),
		VoidedReason:     je.VoidedReason(),
		CreatedAt:        je.CreatedAt(),
		CreatedBy:        je.CreatedBy(),
		UpdatedAt:        je.UpdatedAt(),
		UpdatedBy:        je.UpdatedBy(),
	}
}

func Register(e EchoRouter, commandHandlers *command.Handler) error {
	handler := NewHandler(commandHandlers)

	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}
