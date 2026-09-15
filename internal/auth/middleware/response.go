package middleware

import (
	"log/slog"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := apierror.Wrap(err, apierror.ErrInternal)
	if apiErr.Status >= http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "Внутренняя ошибка middleware аутентификации",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Any("error", apierror.Cause(apiErr)),
		)
	}

	if writeErr := httpx.WriteJSON(w, apiErr.Status, apiErr); writeErr != nil {
		slog.ErrorContext(r.Context(), "Ошибка отправки ответа middleware аутентификации",
			slog.Any("error", writeErr),
		)
	}
}
