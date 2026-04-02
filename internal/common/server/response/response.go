package response

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Envelop map[string]any

func WriteJSON(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data Envelop,
	headers http.Header,
) {
	requestID := middleware.GetReqID(r.Context())

	if len(headers) > 0 {
		for key, value := range headers {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := Envelop{
		"requestID": requestID,
	}
	if data != nil {
		resp["data"] = data
	}

	render.JSON(w, r, resp)
}
