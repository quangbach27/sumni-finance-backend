package ports

import (
	"encoding/json"
	"net/http"
	"sumni-finance-backend/internal/common/server/httperr"
	"sumni-finance-backend/internal/common/server/response"
	"sumni-finance-backend/internal/finance/app/command"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Allocate funds to a wallet
// (POST /v1/wallets/{walletId}/allocate-fund-providers)
func (hs HttpServer) AllocateFund(
	w http.ResponseWriter,
	r *http.Request,
	walletId openapi_types.UUID,
) {
	var req AllocateFundRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		httperr.BadRequest("failed-to-parse-json", err, w, r)
		return
	}

	// Convert request to command
	providers := make([]command.AllocatedProvider, 0, len(req.Providers))
	for _, p := range req.Providers {
		providers = append(providers, command.AllocatedProvider{
			ID:              p.Id,
			AllocatedAmount: p.AllocatedAmount,
		})
	}

	err := hs.application.Commands.AllocateFund.Handle(
		r.Context(),
		command.AllocateFundCmd{
			WalletID:            walletId,
			AllocationProviders: providers,
		},
	)
	if err != nil {
		httperr.RespondWithSlugError(err, w, r)
		return
	}

	response.WriteJSON(w, r, http.StatusCreated, nil, nil)
}
