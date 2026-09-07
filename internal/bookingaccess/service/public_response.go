package service

import (
	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
)

func publicOrderSummary(booking *bookingaccess.Booking) *bookingaccess.PublicSummary {
	return &bookingaccess.PublicSummary{
		Kind:           "order",
		OrderNumber:    booking.OrderNumber,
		Transport:      bookingTransport(booking),
		TicketCount:    len(booking.Tickets),
		PassengerCount: booking.PassengerCount,
		Directions:     bookingDirections(booking),
	}
}

func publicTicketSummary(
	booking *bookingaccess.Booking,
	ticketNumber string,
) (*bookingaccess.PublicSummary, error) {
	for _, ticket := range booking.Tickets {
		if ticket.Number != ticketNumber {
			continue
		}
		return &bookingaccess.PublicSummary{
			Kind:           "ticket",
			OrderNumber:    booking.OrderNumber,
			Transport:      ticket.Transport,
			TicketCount:    len(booking.Tickets),
			PassengerCount: booking.PassengerCount,
			Directions:     []bookingaccess.Direction{ticketDirection(ticket)},
			Ticket: &bookingaccess.PublicTicket{
				TicketNumber: ticket.Number,
				Passenger:    maskedPassenger(ticket.Passenger.FirstName, ticket.Passenger.LastName),
				Segments:     publicSegments(ticket.Segments),
			},
		}, nil
	}
	return nil, apierror.ErrNotFound
}

func bookingTransport(booking *bookingaccess.Booking) string {
	if len(booking.Tickets) == 0 {
		return ""
	}
	return booking.Tickets[0].Transport
}
