package httpx

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
)

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := json.Marshal(data)
	if err != nil {
		slog.Error("Ошибка при формировании JSON", slog.Any("error", err))
		internalErr := apierror.ErrInternal
		w.WriteHeader(internalErr.Status)
		resp, _ = json.Marshal(internalErr)
	} else {
		w.WriteHeader(status)
	}

	if _, err := w.Write(resp); err != nil {
		return fmt.Errorf("write JSON response: %w", err)
	}

	return nil
}
