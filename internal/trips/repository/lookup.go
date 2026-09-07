package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/trips"
)

const tripColumns = `
	SELECT st.transport_type, st.carrier, st.carrier_code, st.route_number,
		st.from_city, st.to_city, st.departure_time, st.arrival_time
	FROM scheduled_trips st
`

func (r *Repository) ByTicketNumber(ctx context.Context, ticketNumber string) ([]trips.ScheduledTrip, error) {
	rows, err := r.db.QueryContext(ctx, tripColumns+`
		JOIN ticket_segments ts ON ts.scheduled_trip_id = st.id
		JOIN tickets t ON t.id = ts.ticket_id
		JOIN orders o ON o.id = t.order_id
		WHERE t.ticket_number = $1 AND o.status IN ($2, $3, $4)
		ORDER BY ts.segment_order
	`, ticketNumber, orders.OrderStatusPaid, orders.OrderStatusPartiallyRefunded, orders.OrderStatusRefunded)
	if err != nil {
		return nil, fmt.Errorf("lookup trips by ticket number: %w", err)
	}
	return scanTrips(rows)
}

func (r *Repository) ByRouteNumber(
	ctx context.Context,
	routeNumber string,
	from time.Time,
	to time.Time,
) ([]trips.ScheduledTrip, error) {
	rows, err := r.db.QueryContext(ctx, tripColumns+`
		WHERE st.route_number = $1 AND st.departure_time >= $2 AND st.departure_time < $3
			AND EXISTS (
				SELECT 1
				FROM ticket_segments ts
				JOIN tickets t ON t.id = ts.ticket_id
				JOIN orders o ON o.id = t.order_id
				WHERE ts.scheduled_trip_id = st.id AND o.status IN ($4, $5, $6)
			)
		ORDER BY st.departure_time, st.id
	`, routeNumber, from, to, orders.OrderStatusPaid,
		orders.OrderStatusPartiallyRefunded, orders.OrderStatusRefunded)
	if err != nil {
		return nil, fmt.Errorf("lookup trips by route number: %w", err)
	}
	return scanTrips(rows)
}

func scanTrips(rows *sql.Rows) ([]trips.ScheduledTrip, error) {
	defer rows.Close()
	result := make([]trips.ScheduledTrip, 0)
	for rows.Next() {
		var item trips.ScheduledTrip
		var departure time.Time
		var arrival time.Time
		if err := rows.Scan(
			&item.Transport, &item.Carrier, &item.CarrierCode, &item.RouteNumber,
			&item.FromCity, &item.ToCity, &departure, &arrival,
		); err != nil {
			return nil, fmt.Errorf("scan scheduled trip: %w", err)
		}
		item.DepartureTime = departure.UTC().Format(time.RFC3339)
		item.ArrivalTime = arrival.UTC().Format(time.RFC3339)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scheduled trips: %w", err)
	}
	if len(result) == 0 {
		return nil, trips.ErrNotFound
	}
	return result, nil
}
