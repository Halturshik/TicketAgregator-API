package api

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/logger"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := json.Marshal(data)
	if err != nil {
		logger.Error("Ошибка при формировании JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		resp = []byte(`{"error":"ошибка при формировании JSON"}`)
	} else {
		w.WriteHeader(status)
	}

	if _, err := w.Write(resp); err != nil {
		logger.Error("Ошибка при отправке ответа: %v", err)
	}
}
