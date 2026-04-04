package ports

import (
	"encoding/json"
	"net/http"
	"sumni-finance-backend/internal/common/server/httperr"
	"sumni-finance-backend/internal/common/server/response"
	"sumni-finance-backend/internal/finance/app/command"
)

// Create a new fund provider
// (POST /v1/fund-providers)
func (hs HttpServer) CreateFundProvider(w http.ResponseWriter, r *http.Request) {
	var req CreateFundProviderRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httperr.BadRequest("failed-to-parse-json", err, w, r)
		return
	}

	if err := hs.application.Commands.CreateFundProvider.Handle(r.Context(), command.CreateFundProviderCmd{
		Name:         req.Name,
		FpType:       req.FpType,
		InitBalance:  req.InitBalance,
		CurrencyCode: req.Currency,
	}); err != nil {
		httperr.RespondWithSlugError(err, w, r)
		return
	}

	response.WriteJSON(w, r, http.StatusCreated, nil, nil)
}
