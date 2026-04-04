package ports

import (
	"encoding/json"
	"net/http"
	"sumni-finance-backend/internal/common/server/httperr"
	"sumni-finance-backend/internal/common/server/response"
	"sumni-finance-backend/internal/finance/app/command"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Open a new accounting period for a wallet
// (POST /v1/wallets/{walletId}/accounting-periods)
func (hs HttpServer) OpenAccountingPeriod(
	w http.ResponseWriter,
	r *http.Request,
	walletId openapi_types.UUID,
) {
	var req OpenAccountingPeriodRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httperr.BadRequest("failed-to-parse-json", err, w, r)
		return
	}

	err := hs.application.Commands.OpenAccountingPeriod.Handle(
		r.Context(),
		command.OpenAccountingPeriodCmd{
			WalletID: walletId,
			Year:     req.Year,
			Month:    req.Month,
		},
	)
	if err != nil {
		httperr.RespondWithSlugError(err, w, r)
		return
	}

	response.WriteJSON(w, r, http.StatusCreated, nil, nil)
}
