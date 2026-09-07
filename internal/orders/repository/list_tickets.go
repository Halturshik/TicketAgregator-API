package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/lib/pq"
)

func (r *Repository) loadHistoryTickets(ctx context.Context, items []orders.HistoryOrder) error {
	orderIDs := make([]int, len(items))
	orderIndexes := make(map[int]int, len(items))
	for index := range items {
		orderIDs[index] = items[index].ID
		orderIndexes[items[index].ID] = index
	}

	rows, err := r.DB.QueryContext(ctx, `
		SELECT t.order_id, t.id, t.ticket_number, t.status, t.transport_type,
			op.passenger_snapshot, op.document_snapshot,
			ts.segment_order, st.from_city, st.to_city, st.departure_time, st.arrival_time
		FROM tickets t
		JOIN order_passengers op ON op.id = t.order_passenger_id AND op.order_id = t.order_id
		JOIN ticket_segments ts ON ts.ticket_id = t.id
		JOIN scheduled_trips st ON st.id = ts.scheduled_trip_id
		WHERE t.order_id = ANY($1)
		ORDER BY t.order_id, t.id, ts.segment_order
	`, pq.Array(orderIDs))
	if err != nil {
		return fmt.Errorf("query order history tickets: %w", err)
	}
	defer rows.Close()

	ticketIndexes := make(map[int]historyTicketIndex)
	for rows.Next() {
		if err := scanHistoryTicketRow(rows, items, orderIndexes, ticketIndexes); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate order history tickets: %w", err)
	}
	buildHistoryDirections(items)
	return nil
}

type historyTicketIndex struct {
	order  int
	ticket int
}

type historyTicketRow struct {
	orderID       int
	ticketID      int
	ticketNumber  string
	status        string
	transport     string
	passengerJSON []byte
	documentJSON  []byte
	segmentOrder  int
	fromCity      string
	toCity        string
	departureTime time.Time
	arrivalTime   time.Time
}

func scanHistoryTicketRow(
	rows interface{ Scan(...any) error },
	items []orders.HistoryOrder,
	orderIndexes map[int]int,
	ticketIndexes map[int]historyTicketIndex,
) error {
	var row historyTicketRow
	if err := rows.Scan(
		&row.orderID, &row.ticketID, &row.ticketNumber, &row.status, &row.transport,
		&row.passengerJSON, &row.documentJSON, &row.segmentOrder,
		&row.fromCity, &row.toCity, &row.departureTime, &row.arrivalTime,
	); err != nil {
		return fmt.Errorf("scan order history ticket: %w", err)
	}

	index, exists := ticketIndexes[row.ticketID]
	if !exists {
		orderIndex, ok := orderIndexes[row.orderID]
		if !ok {
			return fmt.Errorf("order history ticket references unknown order %d", row.orderID)
		}
		ticket, err := historyTicket(row)
		if err != nil {
			return err
		}
		items[orderIndex].Tickets = append(items[orderIndex].Tickets, ticket)
		index = historyTicketIndex{order: orderIndex, ticket: len(items[orderIndex].Tickets) - 1}
		ticketIndexes[row.ticketID] = index
	}
	updateHistoryDirection(&items[index.order].Tickets[index.ticket].Direction, row)
	return nil
}

func historyTicket(row historyTicketRow) (orders.HistoryTicket, error) {
	var passenger orders.PassengerSnapshot
	if err := json.Unmarshal(row.passengerJSON, &passenger); err != nil {
		return orders.HistoryTicket{}, fmt.Errorf("unmarshal order history passenger: %w", err)
	}
	var document orders.DocumentSnapshot
	if err := json.Unmarshal(row.documentJSON, &document); err != nil {
		return orders.HistoryTicket{}, fmt.Errorf("unmarshal order history document: %w", err)
	}
	return orders.HistoryTicket{
		ID: row.ticketID, TicketNumber: row.ticketNumber, Status: row.status,
		Transport: row.transport,
		Passenger: orders.HistoryPassenger{
			FirstName: passenger.FirstName, MiddleName: passenger.MiddleName, LastName: passenger.LastName,
		},
		Document: orders.HistoryDocument{Type: document.Type, Number: document.Number},
	}, nil
}
