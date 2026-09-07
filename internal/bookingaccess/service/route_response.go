package service

import (
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
)

func bookingDirections(booking *bookingaccess.Booking) []bookingaccess.Direction {
	result := make([]bookingaccess.Direction, 0)
	seen := make(map[string]struct{})
	for _, ticket := range booking.Tickets {
		if len(ticket.Segments) == 0 {
			continue
		}
		direction := ticketDirection(ticket)
		key := fmt.Sprintf(
			"%s\x00%s\x00%s\x00%s",
			direction.FromCity,
			direction.ToCity,
			direction.DepartureTime,
			direction.ArrivalTime,
		)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, direction)
	}
	return result
}

func ticketDirection(ticket bookingaccess.Ticket) bookingaccess.Direction {
	if len(ticket.Segments) == 0 {
		return bookingaccess.Direction{}
	}
	first := ticket.Segments[0]
	last := ticket.Segments[len(ticket.Segments)-1]
	return bookingaccess.Direction{
		FromCity:      first.FromCity,
		ToCity:        last.ToCity,
		DepartureTime: first.DepartureTime.UTC().Format(time.RFC3339),
		ArrivalTime:   last.ArrivalTime.UTC().Format(time.RFC3339),
		TransferCount: len(ticket.Segments) - 1,
	}
}

func publicSegments(segments []bookingaccess.Segment) []bookingaccess.PublicSegment {
	result := make([]bookingaccess.PublicSegment, 0, len(segments))
	for _, segment := range segments {
		result = append(result, bookingaccess.PublicSegment{
			Carrier:       segment.Carrier,
			CarrierCode:   segment.CarrierCode,
			RouteNumber:   segment.RouteNumber,
			FromCity:      segment.FromCity,
			ToCity:        segment.ToCity,
			DepartureTime: segment.DepartureTime.UTC().Format(time.RFC3339),
			ArrivalTime:   segment.ArrivalTime.UTC().Format(time.RFC3339),
		})
	}
	return result
}
