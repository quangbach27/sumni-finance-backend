package ports

import (
	"encoding/json"
	"net/http"
	"sumni-finance-backend/internal/common/server/httperr"
	"sumni-finance-backend/internal/common/server/response"
	"sumni-finance-backend/internal/finance/app/command"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Record transaction records for an accounting period
// (POST /v1/wallets/{walletId}/account-periods/{accountPeriodId})
func (hs HttpServer) RecordTransactionRecords(
	w http.ResponseWriter,
	r *http.Request,
	walletId openapi_types.UUID,
	accountPeriodId openapi_types.UUID,
) {
	var req RecordTransactionRecordsRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httperr.BadRequest("failed-to-parse-json", err, w, r)
		return
	}

	// Convert request to command
	transactionRecords := make([]command.TransactionRecordCmd, 0, len(req.TransactionRecords))
	for _, tr := range req.TransactionRecords {
		transactionRecords = append(transactionRecords, command.TransactionRecordCmd{
			FundProviderID:  tr.FundProviderId,
			Amount:          tr.Amount,
			TransactionNo:   tr.TransactionNo,
			TransactionType: tr.TransactionType,
			Description:     tr.Description,
		})
	}

	err := hs.application.Commands.RecordTransactionRecords.Handle(
		r.Context(),
		command.RecordTransactionRecordsCmd{
			WalletID:           walletId,
			AccountingPeridID:  accountPeriodId,
			YearMonth:          req.YearMonth,
			TransactionRecords: transactionRecords,
		},
	)
	if err != nil {
		httperr.RespondWithSlugError(err, w, r)
		return
	}

	response.WriteJSON(w, r, http.StatusCreated, nil, nil)
}
