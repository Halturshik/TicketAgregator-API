package middleware

import (
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := apierror.Wrap(err, apierror.ErrInternal)
	if apiErr.Status >= http.StatusInternalServerError {
		logger.Error("Внутренняя ошибка middleware аутентификации: %s %s: %v", r.Method, r.URL.Path, err)
	} else {
		logger.Warn("Запрос отклонен middleware аутентификации: code=%s method=%s path=%s", apiErr.Code, r.Method, r.URL.Path)
	}

	if writeErr := httpx.WriteJSON(w, apiErr.Status, apiErr); writeErr != nil {
		logger.Error("Ошибка отправки ответа middleware аутентификации: %v", writeErr)
	}
}
