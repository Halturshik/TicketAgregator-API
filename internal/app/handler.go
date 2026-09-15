package app

import (
	"log/slog"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
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
	apiErr := apierror.Wrap(err, apierror.ErrInternal)

	if apiErr.Status >= 500 {
		slog.ErrorContext(r.Context(), "Внутренняя ошибка HTTP API",
			slog.String("error_code", apiErr.Code),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", apierror.Cause(apiErr)),
		)
	}

	if writeErr := httpx.WriteJSON(w, apiErr.Status, apiErr); writeErr != nil {
		slog.ErrorContext(r.Context(), "Ошибка отправки ответа HTTP API", slog.Any("error", writeErr))
	}
}
