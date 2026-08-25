package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func insertTickets(
	ctx context.Context,
	tx *sql.Tx,
	orderID int,
	orderPassengerIDs []int,
	params orders.CreateOrderParams,
) ([]orders.Ticket, error) {
	tickets := make([]orders.Ticket, 0, len(params.Tickets))
	for _, draft := range params.Tickets {
		if draft.PassengerIndex < 0 || draft.PassengerIndex >= len(params.Passengers) {
			return nil, fmt.Errorf("invalid passenger index %d", draft.PassengerIndex)
		}
		passenger := params.Passengers[draft.PassengerIndex]
		ticket, err := insertTicket(ctx, tx, orderID, orderPassengerIDs[draft.PassengerIndex], draft, passenger)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, *ticket)
	}
	return tickets, nil
}

func insertTicket(
	ctx context.Context,
	tx *sql.Tx,
	orderID int,
	orderPassengerID int,
	draft orders.TicketDraft,
	passenger orders.OrderPassengerDraft,
) (*orders.Ticket, error) {
	var ticketID int
	err := tx.QueryRowContext(ctx, `
		INSERT INTO tickets
			(order_id, order_passenger_id, ticket_number, transport_type, is_international, price)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, orderID, orderPassengerID, draft.TicketNumber, draft.Transport, draft.IsInternational, draft.Price).Scan(&ticketID)
	if err != nil {
		return nil, fmt.Errorf("insert ticket: %w", err)
	}
	if err := insertTicketSegments(ctx, tx, ticketID, draft); err != nil {
		return nil, err
	}

	return &orders.Ticket{
		ID: ticketID, TicketNumber: draft.TicketNumber, Transport: draft.Transport,
		IsInternational: draft.IsInternational, Price: draft.Price, Status: orders.TicketStatusBooked,
		Passenger: passenger.Passenger, Document: passenger.Document, Segments: draft.Segments,
	}, nil
}

func insertTicketSegments(ctx context.Context, tx *sql.Tx, ticketID int, draft orders.TicketDraft) error {
	for _, segment := range draft.Segments {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO ticket_segments
				(ticket_id, segment_order, from_city_id, to_city_id, from_city, to_city, departure_time, arrival_time, carrier_id, carrier, carrier_code, route_number)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9, 0),$10,$11,$12)
		`, ticketID, segment.Order, segment.FromCityID, segment.ToCityID, segment.FromCity, segment.ToCity, segment.DepartureTime, segment.ArrivalTime, segment.CarrierID, segment.Carrier, segment.CarrierCode, segment.RouteNumber)
		if err != nil {
			return fmt.Errorf("insert ticket segment: %w", err)
		}
	}
	return nil
}
