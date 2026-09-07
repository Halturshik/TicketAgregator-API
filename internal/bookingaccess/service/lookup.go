package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
)

func (s *Service) Lookup(ctx context.Context, input bookingaccess.LookupInput) (*bookingaccess.PublicSummary, error) {
	location, err := locator(input.TicketNumber, input.OrderNumber)
	if err != nil {
		return nil, err
	}

	var booking *bookingaccess.Booking
	if location.TicketNumber != "" {
		booking, err = s.repo.PublicByTicket(ctx, location.TicketNumber)
	} else {
		booking, err = s.repo.PublicByOrder(ctx, location.OrderNumber)
	}
	if err != nil {
		return nil, mapError("публичного поиска бронирования", err)
	}

	if location.TicketNumber != "" {
		return publicTicketSummary(booking, location.TicketNumber)
	}
	return publicOrderSummary(booking), nil
}
