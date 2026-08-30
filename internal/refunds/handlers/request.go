package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/go-chi/chi/v5"
)

func parseRequest(r *http.Request) (int, *int, refunds.Input, error) {
	orderID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || orderID <= 0 {
		return 0, nil, refunds.Input{}, apierror.ErrInvalidRequest
	}
	var input refunds.Input
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return 0, nil, refunds.Input{}, apierror.ErrInvalidJSON
	}
	var userID *int
	if id, ok := auth.UserIDFromContext(r.Context()); ok {
		userID = &id
	}
	return orderID, userID, input, nil
}
