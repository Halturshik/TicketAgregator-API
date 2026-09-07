package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
	scheduledTripIDs := make(map[string]int)
	for _, draft := range params.Tickets {
		if draft.PassengerIndex < 0 || draft.PassengerIndex >= len(params.Passengers) {
			return nil, fmt.Errorf("invalid passenger index %d", draft.PassengerIndex)
		}
		passenger := params.Passengers[draft.PassengerIndex]
		ticket, err := insertTicket(
			ctx,
			tx,
			orderID,
			orderPassengerIDs[draft.PassengerIndex],
			draft,
			passenger,
			scheduledTripIDs,
		)
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
	scheduledTripIDs map[string]int,
) (*orders.Ticket, error) {
	policy, err := json.Marshal(draft.RefundPolicy)
	if err != nil {
		return nil, fmt.Errorf("marshal refund policy: %w", err)
	}
	var ticketID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO tickets
			(order_id, order_passenger_id, ticket_number, supplier_code, supplier_offer_id,
			 fare_type, refund_policy_version, refund_policy_snapshot,
			 transport_type, is_international, price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, orderID, orderPassengerID, draft.TicketNumber, draft.SupplierCode, draft.SupplierOfferID,
		draft.FareType, draft.RefundPolicyVersion, policy, draft.Transport, draft.IsInternational,
		draft.Price).Scan(&ticketID)
	if err != nil {
		return nil, fmt.Errorf("insert ticket: %w", err)
	}
	if err := insertTicketSegments(ctx, tx, ticketID, draft, scheduledTripIDs); err != nil {
		return nil, err
	}

	return &orders.Ticket{
		ID: ticketID, TicketNumber: draft.TicketNumber, Transport: draft.Transport,
		SupplierCode: draft.SupplierCode, SupplierOfferID: draft.SupplierOfferID,
		FareType: draft.FareType, RefundPolicy: draft.RefundPolicy, RefundPolicyVersion: draft.RefundPolicyVersion,
		IsInternational: draft.IsInternational, Price: draft.Price,
		Status:    orders.TicketStatusBooked,
		Passenger: passenger.Passenger, Document: passenger.Document, Segments: draft.Segments,
	}, nil
}

func insertTicketSegments(
	ctx context.Context,
	tx *sql.Tx,
	ticketID int,
	draft orders.TicketDraft,
	scheduledTripIDs map[string]int,
) error {
	for _, segment := range draft.Segments {
		tripKey := scheduledTripKey(draft.Transport, segment)
		scheduledTripID, exists := scheduledTripIDs[tripKey]
		if !exists {
			var err error
			scheduledTripID, err = insertScheduledTrip(ctx, tx, tripKey, draft.Transport, segment)
			if err != nil {
				return err
			}
			scheduledTripIDs[tripKey] = scheduledTripID
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO ticket_segments
				(ticket_id, scheduled_trip_id, segment_order)
			VALUES ($1, $2, $3)
		`, ticketID, scheduledTripID, segment.Order)
		if err != nil {
			return fmt.Errorf("insert ticket segment: %w", err)
		}
	}
	return nil
}
