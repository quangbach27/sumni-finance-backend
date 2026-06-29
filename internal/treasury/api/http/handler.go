package http

import (
	"context"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/treasury/app/command"
	"sumni-finance-backend/internal/treasury/app/query"
	"sumni-finance-backend/internal/treasury/domain"
)

type Handler struct {
	commandHandlers *command.Handlers
	queryHandlers   *query.Handlers
}

func NewHandler(
	commandHandlers *command.Handlers,
	queryHandlers *query.Handlers,
) Handler {
	if commandHandlers == nil {
		panic("command handler can't be nil")
	}

	if queryHandlers == nil {
		panic("query handler can't be nil")
	}

	return Handler{
		commandHandlers: commandHandlers,
		queryHandlers:   queryHandlers,
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

	fundSourceUUID, err := h.commandHandlers.RegisterFundSource(ctx, command.RegisterFundSourceCmd{
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

	return RegisterFundSource201JSONResponse{
		FundSourceUuid: fundSourceUUID,
	}, nil
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

	err := h.commandHandlers.RecordJournalEntries(ctx, command.RecordedJournalEntriesCmd{
		FundSourceUUID:    request.FundSourceUuid,
		JournalEntryItems: journalEntryItems,
	})
	if err != nil {
		return nil, err
	}

	return RecordJournalEntries201JSONResponse{}, nil
}

// Void a journal entry by creating a reverse journal entry
// (POST /v1/treasury/fund-sources/{fund_source_uuid}/journal-entries/{journal_entry_uuid}/void)
func (h Handler) VoidJournalEntry(ctx context.Context, request VoidJournalEntryRequestObject) (VoidJournalEntryResponseObject, error) {
	reverseJournalEntry, err := h.commandHandlers.VoidJournalEntry(ctx, command.VoidJournalEntryCmd{
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
		CreatedAt:        je.Audit().CreatedAt(),
		CreatedBy:        je.Audit().CreatedBy(),
		UpdatedAt:        je.Audit().UpdatedAt(),
		UpdatedBy:        je.Audit().UpdatedBy(),
	}
}

// List all fund sources
// (GET /v1/treasury/fund-sources)
func (h Handler) ListFundSources(ctx context.Context, request ListFundSourcesRequestObject) (ListFundSourcesResponseObject, error) {
	views, err := h.queryHandlers.ListFundSources(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]FundSourceResponse, len(views))
	for i, v := range views {
		items[i] = fundSourceViewToResponse(v)
	}

	return ListFundSources200JSONResponse{
		Items: items,
	}, nil
}

func fundSourceViewToResponse(v query.FundSourceView) FundSourceResponse {
	r := FundSourceResponse{
		FundSourceUuid:   v.UUID,
		Name:             v.Name,
		SourceType:       v.SourceType,
		Balance:          v.Balance,
		AvailableBalance: v.AvailableBalance,
		Currency:         v.Currency,
		CreatedAt:        v.CreatedAt,
		CreatedBy:        v.CreatedBy,
		UpdatedAt:        v.UpdatedAt,
		UpdatedBy:        v.UpdatedBy,
	}

	if v.BankMetadata != nil {
		r.BankMetadata = &FundSourceBankMetadataResponse{
			BankName:      v.BankMetadata.BankName,
			BankCode:      v.BankMetadata.BankCode,
			BankShortName: v.BankMetadata.BankShortName,
			BankBin:       v.BankMetadata.BankBin,
			BankLogoUrl:   v.BankMetadata.BankLogoUrl,
			BankIconUrl:   v.BankMetadata.BankIconUrl,
			AccountNumber: v.BankMetadata.AccountNumber,
			AccountOwner:  v.BankMetadata.AccountOwner,
		}
	}

	if v.CashMetadata != nil {
		r.CashMetadata = &FundSourceCashMetadataResponse{
			OwnerName: v.CashMetadata.OwnerName,
		}
	}

	return r
}

const (
	defaultPage     = 1
	defaultPageSize = 20
)

// List journal entries for a fund source
// (GET /v1/treasury/fund-sources/{fund_source_uuid}/journal-entries)
func (h Handler) ListJournalEntries(ctx context.Context, request ListJournalEntriesRequestObject) (ListJournalEntriesResponseObject, error) {
	page := defaultPage
	if request.Params.Page != nil {
		page = *request.Params.Page
	}
	pageSize := defaultPageSize
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}

	result, err := h.queryHandlers.ListJournalEntries(ctx, query.ListJournalEntriesQuery{
		FundSourceUUID: request.FundSourceUuid,
		Page:           page,
		PageSize:       pageSize,
		DateFrom:       request.Params.DateFrom,
		DateTo:         request.Params.DateTo,
	})
	if err != nil {
		return nil, err
	}

	items := make([]JournalEntryResponse, len(result.Items))
	for i, v := range result.Items {
		items[i] = journalEntryViewToResponse(v)
	}

	return ListJournalEntries200JSONResponse{
		Items:      items,
		TotalItems: result.TotalItems,
		Page:       result.Page,
		PageSize:   result.PageSize,
	}, nil
}

func journalEntryViewToResponse(v query.JournalEntryView) JournalEntryResponse {
	return JournalEntryResponse{
		JournalEntryUuid: v.UUID,
		FundSourceUuid:   v.FundSourceUUID,
		Amount:           v.Amount,
		EntryType:        v.EntryType,
		TransactionDate:  v.TransactionDate,
		TransactionNo:    v.TransactionNo,
		Description:      v.Description,
		Status:           v.Status,
		BalanceBefore:    v.BalanceBefore,
		BalanceAfter:     v.BalanceAfter,
		ReverseEntryUuid: v.ReverseEntryUUID,
		VoidedBy:         v.VoidedBy,
		VoidedAt:         v.VoidedAt,
		VoidedReason:     v.VoidedReason,
		CreatedAt:        v.CreatedAt,
		CreatedBy:        v.CreatedBy,
		UpdatedAt:        v.UpdatedAt,
		UpdatedBy:        v.UpdatedBy,
	}
}

func Register(
	protectedRouter common.EchoRouter,
	commandHandlers *command.Handlers,
	queryHandlers *query.Handlers,
) error {
	handler := NewHandler(commandHandlers, queryHandlers)

	RegisterHandlers(protectedRouter, NewStrictHandler(handler, nil))

	return nil
}
