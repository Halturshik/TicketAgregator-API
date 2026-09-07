package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) QuoteRefund(w http.ResponseWriter, r *http.Request) error {
	input, err := decode[bookingaccess.RefundInput](r)
	if err != nil {
		return err
	}
	userID, token := accessCredentials(r)
	output, err := h.service.QuoteRefund(
		r.Context(), userID, chi.URLParam(r, "orderNumber"), token, input,
	)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	w.Header().Set("Cache-Control", "no-store")
	return httpx.WriteJSON(w, http.StatusOK, output)
}

func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) error {
	input, err := decode[bookingaccess.RefundInput](r)
	if err != nil {
		return err
	}
	userID, token := accessCredentials(r)
	output, err := h.service.Refund(
		r.Context(), userID, chi.URLParam(r, "orderNumber"), token,
		input, r.Header.Get("Idempotency-Key"),
	)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	w.Header().Set("Cache-Control", "no-store")
	status := http.StatusCreated
	if output.Status == refunds.StatusProcessing {
		status = http.StatusAccepted
		setRetryAfter(w, output.NextRetryAt)
	} else if output.Status == refunds.StatusRequiresReview {
		status = http.StatusAccepted
	} else if output.Status == refunds.StatusFailed {
		status = http.StatusConflict
	}
	return httpx.WriteJSON(w, status, output)
}

func setRetryAfter(w http.ResponseWriter, nextRetryAt *time.Time) {
	if nextRetryAt == nil {
		return
	}
	delay := time.Until(*nextRetryAt)
	retryAfter := int((delay + time.Second - 1) / time.Second)
	if retryAfter < 1 {
		retryAfter = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
}
