package api

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/logger"
)

func (api *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FirstName  string `json:"first_name"`
		MiddleName string `json:"middle_name,omitempty"`
		LastName   string `json:"last_name"`
		BirthDate  string `json:"birth_date"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		IsRussian  bool   `json:"is_russian"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Ошибка: не удалось прочитать тело запроса: %v", err)
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Некорректно оформлено тело запроса"})
		return
	}
}
