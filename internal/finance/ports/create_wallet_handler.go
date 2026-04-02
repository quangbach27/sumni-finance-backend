package ports

import (
	"encoding/json"
	"net/http"
	"sumni-finance-backend/internal/common/server/httperr"
	"sumni-finance-backend/internal/common/server/response"
	"sumni-finance-backend/internal/finance/app/command"
)

// Create a new wallet
// (POST /v1/wallet)
func (hs HttpServer) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req CreateWalletRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httperr.BadRequest("failed-to-parse-json", err, w, r)
		return
	}

	if err := hs.application.Commands.CreateWallet.Handle(r.Context(), command.CreateWalletCmd{
		Name:         req.Name,
		CurrencyCode: req.Currency,
	}); err != nil {
		httperr.RespondWithSlugError(err, w, r)
		return
	}

	response.WriteJSON(w, r, http.StatusCreated, nil, nil)
}
