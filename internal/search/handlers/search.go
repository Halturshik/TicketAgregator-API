package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (h *Handler) SearchAir(w http.ResponseWriter, r *http.Request) error {
	return h.search(w, r, "avia")
}

func (h *Handler) SearchRailway(w http.ResponseWriter, r *http.Request) error {
	return h.search(w, r, "rail")
}

func (h *Handler) SearchBus(w http.ResponseWriter, r *http.Request) error {
	return h.search(w, r, "bus")
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request, transport string) error {
	var req search.SearchInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierror.ErrInvalidJSON
	}
	userID := optionalUserID(r)
	result, err := h.service.Search(r.Context(), transport, req, userID)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	return httpx.WriteJSON(w, http.StatusOK, result)
}

func optionalUserID(r *http.Request) *int {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		return nil
	}
	return &userID
}
