package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
)

func decode[T any](r *http.Request) (T, error) {
	var input T
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return input, apierror.ErrInvalidJSON
	}
	return input, nil
}

func accessCredentials(r *http.Request) (userID *int, token string) {
	if id, ok := auth.UserIDFromContext(r.Context()); ok {
		userID = &id
	}
	return userID, bookingaccess.AccessToken(r)
}
