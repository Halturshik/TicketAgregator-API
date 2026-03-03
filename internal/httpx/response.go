package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
)

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := json.Marshal(data)
	if err != nil {
		logger.Error("Ошибка при формировании JSON: %v", err)
		internalErr := apierror.ErrInternal
		w.WriteHeader(internalErr.Status)
		resp, _ = json.Marshal(internalErr)
	} else {
		w.WriteHeader(status)
	}

	if _, err := w.Write(resp); err != nil {
		logger.Error("Ошибка при отправке ответа: %v", err)
		return err
	}

	return nil
}
