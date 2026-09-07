package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
)

func bookingDetails(booking *bookingaccess.Booking) *bookingaccess.BookingDetails {
	result := &bookingaccess.BookingDetails{
		OrderNumber:       booking.OrderNumber,
		Status:            booking.Status,
		CurrentTotalPrice: booking.CurrentTotalPrice,
		BonusSpent:        booking.BonusSpent,
		BonusEarned:       booking.BonusEarned,
		CreatedAt:         booking.CreatedAt.UTC().Format(time.RFC3339),
		PassengerCount:    booking.PassengerCount,
		Directions:        bookingDirections(booking),
		Tickets:           []bookingaccess.DetailTicket{},
	}
	for _, ticket := range booking.Tickets {
		result.Tickets = append(result.Tickets, bookingaccess.DetailTicket{
			TicketNumber: ticket.Number,
			Status:       ticket.Status,
			Transport:    ticket.Transport,
			FareType:     ticket.FareType,
			Passenger:    maskedPassenger(ticket.Passenger.FirstName, ticket.Passenger.LastName),
			DocumentType: ticket.Document.Type,
			Document:     maskedDocument(ticket.Document.Number),
			Segments:     publicSegments(ticket.Segments),
		})
	}
	return result
}
