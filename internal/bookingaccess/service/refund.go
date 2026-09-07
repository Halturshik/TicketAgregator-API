package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	ordernumber "github.com/Halturshik/TicketAgregator-API/internal/orders/number"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func (s *Service) QuoteRefund(
	ctx context.Context,
	userID *int,
	orderNumber string,
	accessToken string,
	input bookingaccess.RefundInput,
) (*bookingaccess.RefundQuote, error) {
	booking, refundInput, err := s.prepareRefund(ctx, userID, orderNumber, accessToken, input)
	if err != nil {
		return nil, err
	}
	quote, err := s.refunds.Quote(ctx, userID, booking.ID, refundInput)
	if err != nil {
		return nil, err
	}
	return refundQuote(booking, quote), nil
}

func (s *Service) Refund(
	ctx context.Context,
	userID *int,
	orderNumber string,
	accessToken string,
	input bookingaccess.RefundInput,
	idempotencyKey string,
) (*bookingaccess.RefundResult, error) {
	booking, refundInput, err := s.prepareRefund(ctx, userID, orderNumber, accessToken, input)
	if err != nil {
		return nil, err
	}
	refundInput.IdempotencyKey = idempotencyKey
	result, err := s.refunds.Refund(ctx, userID, booking.ID, refundInput)
	if err != nil {
		return nil, err
	}
	return refundResult(booking, result), nil
}

func (s *Service) prepareRefund(
	ctx context.Context,
	userID *int,
	orderNumber string,
	accessToken string,
	input bookingaccess.RefundInput,
) (*bookingaccess.Booking, refunds.Input, error) {
	orderNumber, err := validOrderNumber(orderNumber)
	if err != nil {
		return nil, refunds.Input{}, err
	}
	booking, err := s.repo.ByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, refunds.Input{}, mapError("получения заказа для возврата", err)
	}
	if err := s.authorize(ctx, booking, userID, accessToken); err != nil {
		return nil, refunds.Input{}, mapError("проверки доступа к возврату", err)
	}
	selection, err := refundSelection(booking, input)
	if err != nil {
		return nil, refunds.Input{}, err
	}
	if booking.UserID == nil {
		selection.GuestPaymentToken = booking.GuestPaymentToken
	}
	return booking, selection, nil
}

func refundSelection(booking *bookingaccess.Booking, input bookingaccess.RefundInput) (refunds.Input, error) {
	if input.All == (len(input.TicketNumbers) > 0) {
		return refunds.Input{}, apierror.ErrInvalidRequest
	}
	if input.All {
		return refunds.Input{All: true}, nil
	}

	available := make(map[string]int, len(booking.Tickets))
	for _, ticket := range booking.Tickets {
		available[ticket.Number] = ticket.ID
	}
	ids := make([]int, 0, len(input.TicketNumbers))
	seen := make(map[string]struct{}, len(input.TicketNumbers))
	for _, raw := range input.TicketNumbers {
		number := normalizeNumber(raw)
		if !ordernumber.ValidTicket(number) {
			return refunds.Input{}, apierror.ErrInvalidRequest
		}
		if _, exists := seen[number]; exists {
			return refunds.Input{}, apierror.ErrInvalidRequest
		}
		id, exists := available[number]
		if !exists {
			return refunds.Input{}, apierror.ErrNotFound
		}
		seen[number] = struct{}{}
		ids = append(ids, id)
	}
	return refunds.Input{TicketIDs: ids}, nil
}
