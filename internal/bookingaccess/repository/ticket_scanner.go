package repository

import (
	"encoding/json"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
)

type bookingTicketRow struct {
	id            int
	number        string
	status        string
	transport     string
	fareType      string
	policyJSON    []byte
	price         int
	passengerJSON []byte
	documentJSON  []byte
	segment       bookingaccess.Segment
}

func scanBookingTicketRow(source interface{ Scan(...any) error }) (bookingTicketRow, error) {
	var row bookingTicketRow
	if err := source.Scan(
		&row.id,
		&row.number,
		&row.status,
		&row.transport,
		&row.fareType,
		&row.policyJSON,
		&row.price,
		&row.passengerJSON,
		&row.documentJSON,
		&row.segment.Order,
		&row.segment.Carrier,
		&row.segment.CarrierCode,
		&row.segment.RouteNumber,
		&row.segment.FromCity,
		&row.segment.ToCity,
		&row.segment.DepartureTime,
		&row.segment.ArrivalTime,
	); err != nil {
		return bookingTicketRow{}, fmt.Errorf("scan booking ticket: %w", err)
	}
	return row, nil
}

func bookingTicket(row bookingTicketRow) (bookingaccess.Ticket, error) {
	ticket := bookingaccess.Ticket{
		ID:        row.id,
		Number:    row.number,
		Status:    row.status,
		Transport: row.transport,
		FareType:  row.fareType,
		Price:     row.price,
		Segments:  []bookingaccess.Segment{},
	}
	if err := json.Unmarshal(row.policyJSON, &ticket.RefundPolicy); err != nil {
		return bookingaccess.Ticket{}, fmt.Errorf("decode booking refund policy: %w", err)
	}
	if err := json.Unmarshal(row.passengerJSON, &ticket.Passenger); err != nil {
		return bookingaccess.Ticket{}, fmt.Errorf("decode booking passenger: %w", err)
	}
	if err := json.Unmarshal(row.documentJSON, &ticket.Document); err != nil {
		return bookingaccess.Ticket{}, fmt.Errorf("decode booking document: %w", err)
	}
	return ticket, nil
}
