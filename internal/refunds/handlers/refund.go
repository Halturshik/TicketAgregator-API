package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) error {
	orderID, userID, input, err := parseRequest(r)
	if err != nil {
		return err
	}
	input.IdempotencyKey = r.Header.Get("Idempotency-Key")
	output, err := h.service.Refund(r.Context(), userID, orderID, input)
	if err != nil {
		return apierror.Wrap(err, apierror.ErrInternal)
	}
	status := http.StatusCreated
	if output.Status == refunds.StatusProcessing {
		status = http.StatusAccepted
		if output.NextRetryAt != nil {
			delay := time.Until(*output.NextRetryAt)
			retryAfter := int((delay + time.Second - 1) / time.Second)
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		}
	} else if output.Status == refunds.StatusRequiresReview {
		status = http.StatusAccepted
	} else if output.Status == refunds.StatusFailed {
		status = http.StatusConflict
	}
	return httpx.WriteJSON(w, status, output)
}
