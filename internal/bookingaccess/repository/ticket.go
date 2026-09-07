package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
)

func (r *Repository) loadTickets(ctx context.Context, booking *bookingaccess.Booking) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.ticket_number, t.status, t.transport_type, t.fare_type,
			t.refund_policy_snapshot, t.price, op.passenger_snapshot, op.document_snapshot,
			ts.segment_order, st.carrier, st.carrier_code, st.route_number,
			st.from_city, st.to_city, st.departure_time, st.arrival_time
		FROM tickets t
		JOIN order_passengers op ON op.id = t.order_passenger_id AND op.order_id = t.order_id
		JOIN ticket_segments ts ON ts.ticket_id = t.id
		JOIN scheduled_trips st ON st.id = ts.scheduled_trip_id
		WHERE t.order_id = $1
		ORDER BY t.id, ts.segment_order
	`, booking.ID)
	if err != nil {
		return fmt.Errorf("query booking tickets: %w", err)
	}
	defer rows.Close()

	indexes := make(map[int]int)
	for rows.Next() {
		row, err := scanBookingTicketRow(rows)
		if err != nil {
			return err
		}

		index, exists := indexes[row.id]
		if !exists {
			ticket, err := bookingTicket(row)
			if err != nil {
				return err
			}
			booking.Tickets = append(booking.Tickets, ticket)
			index = len(booking.Tickets) - 1
			indexes[row.id] = index
		}
		booking.Tickets[index].Segments = append(booking.Tickets[index].Segments, row.segment)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate booking tickets: %w", err)
	}
	if len(booking.Tickets) == 0 {
		return bookingaccess.ErrNotFound
	}
	booking.CreatedAt = booking.CreatedAt.UTC()
	return nil
}
