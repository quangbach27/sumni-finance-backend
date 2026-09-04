package http

import (
	"context"
	"errors"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/app/command"
	"sumni-finance-backend/internal/treasury/app/query"
)

const (
	// TODO: replace with real tenant context from auth middleware
	mockedTenantID  = "tenant-1"
	mockedOfficeID  = "office-1"
	defaultPageSize = 20
)

func mockedTenantContext() shared.TenantContext {
	tc, _ := shared.NewTenantContext(mockedTenantID, mockedOfficeID)
	return tc
}

type Handler struct {
	queryHandler   *query.Handlers
	commandHandler *command.Handlers
}

func NewHandler(
	queryHandler *query.Handlers,
	commandHandler *command.Handlers,
) *Handler {
	if commandHandler == nil {
		panic("commandHandler can't be nil")
	}

	if queryHandler == nil {
		panic("queryHandler can't be nil")
	}

	return &Handler{
		queryHandler:   queryHandler,
		commandHandler: commandHandler,
	}
}

func (h *Handler) ListFundSources(ctx context.Context, request ListFundSourcesRequestObject) (ListFundSourcesResponseObject, error) {
	result, err := h.queryHandler.ListFundSources(ctx, query.ListFundSources{
		Page:          common.SafeDeref(request.Params.Page, 1),
		PageSize:      common.SafeDeref(request.Params.PageSize, defaultPageSize),
		TenantContext: mockedTenantContext(),
	})
	if err != nil {
		return nil, err
	}

	resp := ListFundSources200JSONResponse{
		Pagination: Pagination{
			Total:    result.TotalCount,
			Page:     result.Page,
			PageSize: result.PageSize,
		},
		Items: make([]FundSourceResponse, 0, len(result.Items)),
	}
	for _, item := range result.Items {
		fsResp := FundSourceResponse{
			FundSourceUuid:   item.UUID,
			Name:             item.Name,
			SourceType:       item.SourceType,
			Currency:         item.Currency,
			Balance:          item.Balance,
			AvailableBalance: item.AvailableBalance,
		}
		if item.BankMetadata != nil {
			fsResp.BankMetadata = &FundSourceBankMetadataResponse{
				BankName:      item.BankMetadata.BankName,
				BankShortName: item.BankMetadata.BankShortName,
				BankCode:      item.BankMetadata.BankBin,
				AccountNumber: item.BankMetadata.AccountNumber,
				AccountOwner:  item.BankMetadata.AccountOwner,
			}
		}
		if item.CashMetadata != nil {
			fsResp.CashMetadata = &FundSourceCashMetadataResponse{
				OwnerName: item.CashMetadata.OwnerName,
			}
		}
		resp.Items = append(resp.Items, fsResp)
	}

	return resp, nil
}

func (h *Handler) CreateFundSource(ctx context.Context, request CreateFundSourceRequestObject) (CreateFundSourceResponseObject, error) {
	if request.Body == nil {
		return nil, common.NewInvalidInputError("invalid-create-fund-source-request", "request body is required")
	}

	cmd := command.CreateFundSource{
		Name:          request.Body.Name,
		SourceType:    request.Body.SourceType,
		InitBalance:   request.Body.InitBalance,
		Currency:      request.Body.Currency,
		TenantContext: mockedTenantContext(),
	}

	if request.Body.Metadata.BankMetadata != nil {
		cmd.BankMetadata = command.BankMetadata{
			BankName:      request.Body.Metadata.BankMetadata.BankCode,
			BankBin:       request.Body.Metadata.BankMetadata.BankCode,
			BankShortName: request.Body.Metadata.BankMetadata.BankCode,
			AccountNumber: request.Body.Metadata.BankMetadata.AccountNumber,
			AccountOwner:  request.Body.Metadata.BankMetadata.AccountOwner,
		}
	}

	if request.Body.Metadata.CashMetadata != nil {
		cmd.CashMetadataCmd = command.CashMetadata{
			OwnerName: request.Body.Metadata.CashMetadata.OwnerName,
		}
	}

	fundSourceUUID, err := h.commandHandler.CreateFundSource(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return CreateFundSource201JSONResponse{FundSourceUuid: fundSourceUUID}, nil
}

func (h *Handler) CreateWallet(ctx context.Context, request CreateWalletRequestObject) (CreateWalletResponseObject, error) {
	if request.Body == nil {
		return nil, common.NewInvalidInputError("invalid-create-wallet-request", "request body is required")
	}

	allocations := make([]command.CreateWalletAllocation, 0)
	if request.Body.Allocations != nil {
		allocations = make([]command.CreateWalletAllocation, 0, len(*request.Body.Allocations))
		for _, allocation := range *request.Body.Allocations {
			allocations = append(allocations, command.CreateWalletAllocation{
				FundSourceUUID:   allocation.FundSourceUuid,
				AllocatedBalance: allocation.AllocatedBalance,
			})
		}
	}

	walletUUID, err := h.commandHandler.CreateWallet(ctx, command.CreateWallet{
		Name:          request.Body.Name,
		Currency:      request.Body.Currency,
		Allocations:   allocations,
		TenantContext: mockedTenantContext(),
	})
	if err != nil {
		return nil, err
	}

	return CreateWallet201JSONResponse{WalletUuid: walletUUID}, nil
}

func (h *Handler) LinkFundSources(ctx context.Context, request LinkFundSourcesRequestObject) (LinkFundSourcesResponseObject, error) {
	if request.Body == nil {
		return nil, common.NewInvalidInputError("invalid-link-fund-sources-request", "request body is required")
	}

	allocations := make([]command.LinkFundSourcesAllocation, 0, len(request.Body.Allocations))
	for _, allocation := range request.Body.Allocations {
		allocations = append(allocations, command.LinkFundSourcesAllocation{
			FundSourceUUID:  allocation.FundSourceUuid,
			AllocatedAmount: allocation.AllocatedAmount,
		})
	}

	err := h.commandHandler.LinkFundSources(ctx, command.LinkFundSources{
		WalletUUID:  request.WalletUuid,
		Allocations: allocations,
		Tenant:      mockedTenantContext(),
	})
	if err != nil {
		return nil, err
	}

	return LinkFundSources204Response{}, nil
}

func (h *Handler) ListWallets(ctx context.Context, request ListWalletsRequestObject) (ListWalletsResponseObject, error) {
	result, err := h.queryHandler.ListWallets(ctx, query.ListWallets{
		Page:          common.SafeDeref(request.Params.Page, 1),
		PageSize:      common.SafeDeref(request.Params.PageSize, defaultPageSize),
		TenantContext: mockedTenantContext(),
	})
	if err != nil {
		return nil, err
	}

	resp := ListWallets200JSONResponse{
		Pagination: Pagination{
			Total:    result.TotalCount,
			Page:     result.Page,
			PageSize: result.PageSize,
		},
		Items: make([]WalletResponse, 0, len(result.Items)),
	}
	for _, item := range result.Items {
		allocations := make([]WalletAllocationResponse, 0, len(item.Allocations))
		for _, a := range item.Allocations {
			allocations = append(allocations, WalletAllocationResponse{
				FundSourceUuid: a.FundSourceUUID,
				FundSourceName: a.FundSourceName,
				Balance:        a.Balance,
			})
		}
		resp.Items = append(resp.Items, WalletResponse{
			WalletUuid:  item.UUID,
			Name:        item.Name,
			Currency:    item.Currency,
			Balance:     item.Balance,
			Allocations: allocations,
		})
	}

	return resp, nil
}

func (h *Handler) CreateTransactions(ctx context.Context, request CreateTransactionsRequestObject) (CreateTransactionsResponseObject, error) {
	if request.Body == nil {
		return nil, common.NewInvalidInputError("invalid-create-transactions-request", "request body is required")
	}

	items := make([]command.CreateTransactionItem, 0, len(request.Body.Items))
	for _, item := range request.Body.Items {
		items = append(items, command.CreateTransactionItem{
			FundSourceUUID:  item.FundSourceUuid,
			WalletUUID:      item.WalletUuid,
			EntryType:       item.EntryType,
			Amount:          item.Amount,
			Currency:        item.Currency,
			Description:     item.Description,
			TransactionDate: item.TransactionDate,
		})
	}

	err := h.commandHandler.CreateTransactions(ctx, command.CreateTransactions{
		Items:         items,
		TenantContext: mockedTenantContext(),
	})
	if err != nil {
		return nil, err
	}

	return CreateTransactions204Response{}, nil
}

func Register(protectedRoute EchoRouter, handler *Handler) error {
	if handler == nil {
		return errors.New("handler can't be nil")
	}

	RegisterHandlers(protectedRoute, NewStrictHandler(handler, nil))
	return nil
}
