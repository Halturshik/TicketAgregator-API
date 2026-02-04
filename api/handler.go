package api

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
)

type HandlerWithError func(http.ResponseWriter, *http.Request) error

func (api *API) Handle(h HandlerWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		api.handleAPIError(w, r, err)
	}
}

func (api *API) handleAPIError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *apierror.APIError
	if ae, ok := err.(*apierror.APIError); ok {
		apiErr = ae
	} else {
		apiErr = apierror.Wrap(err, apierror.ErrInternal)
	}

	logger.Warn("API error: %s | %s | %s %s",
		apiErr.Code,
		apiErr.Message,
		r.Method,
		r.URL.Path,
	)

	if apiErr.Status >= 500 {
		logger.Error("server error: %v", err)
	}

	writeJSON(w, apiErr.Status, apiErr)
}
