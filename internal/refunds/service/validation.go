package service

import (
	"strings"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/google/uuid"
)

func validateSelection(orderID int, input refunds.Input) error {
	if orderID <= 0 || input.All == (len(input.TicketIDs) > 0) {
		return apierror.ErrInvalidRequest
	}
	seen := make(map[int]struct{}, len(input.TicketIDs))
	for _, ticketID := range input.TicketIDs {
		if ticketID <= 0 {
			return apierror.ErrInvalidRequest
		}
		if _, exists := seen[ticketID]; exists {
			return apierror.ErrInvalidRequest
		}
		seen[ticketID] = struct{}{}
	}
	return nil
}

func validateRefundInput(orderID int, input refunds.Input) error {
	if err := validateSelection(orderID, input); err != nil {
		return err
	}
	if input.IdempotencyKey == "" || input.IdempotencyKey != strings.TrimSpace(input.IdempotencyKey) ||
		uuid.Validate(input.IdempotencyKey) != nil {
		return apierror.ErrInvalidRequest
	}
	return nil
}

func authorize(order *refunds.OrderData, userID *int, guestToken string) error {
	if order.UserID != nil {
		if userID == nil || *order.UserID != *userID {
			return refunds.ErrForbidden
		}
		return nil
	}
	if order.GuestToken == "" || strings.TrimSpace(guestToken) == "" || order.GuestToken != strings.TrimSpace(guestToken) {
		return refunds.ErrForbidden
	}
	return nil
}

func ensureRefundableOrder(order *refunds.OrderData) error {
	if order.Status != orders.OrderStatusPaid && order.Status != orders.OrderStatusPartiallyRefunded {
		return refunds.ErrInvalidStatus
	}
	return nil
}

func ensurePaidTickets(tickets []refunds.TicketData, expected int, all bool) error {
	if len(tickets) == 0 || !all && len(tickets) != expected {
		return refunds.ErrTicketsMismatch
	}
	for _, ticket := range tickets {
		if ticket.Status != orders.TicketStatusPaid {
			return refunds.ErrInvalidStatus
		}
	}
	return nil
}
