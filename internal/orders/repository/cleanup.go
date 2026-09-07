package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (r *Repository) DeleteExpiredOrders(ctx context.Context, before time.Time, limit int) (int, error) {
	result, err := r.DB.ExecContext(ctx, `
		WITH stale_orders AS (
			SELECT id
			FROM orders
			WHERE status IN ($1, $2) AND expires_at <= $3
			ORDER BY expires_at, id
			LIMIT $4
			FOR UPDATE SKIP LOCKED
		)
		DELETE FROM orders o
		USING stale_orders stale
		WHERE o.id = stale.id
	`, orders.OrderStatusCreated, orders.OrderStatusExpired, before, limit)
	if err != nil {
		return 0, fmt.Errorf("delete expired orders: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count deleted expired orders: %w", err)
	}
	return int(deleted), nil
}

func (r *Repository) DeleteOrphanTrips(ctx context.Context, before time.Time, limit int) (int, error) {
	result, err := r.DB.ExecContext(ctx, `
		WITH orphan_trips AS (
			SELECT trip.id
			FROM scheduled_trips trip
			WHERE trip.created_at <= $1
				AND NOT EXISTS (
					SELECT 1 FROM ticket_segments segment
					WHERE segment.scheduled_trip_id = trip.id
				)
			ORDER BY trip.created_at, trip.id
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		DELETE FROM scheduled_trips trip
		USING orphan_trips orphan
		WHERE trip.id = orphan.id
	`, before, limit)
	if err != nil {
		return 0, fmt.Errorf("delete orphan scheduled trips: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count deleted orphan scheduled trips: %w", err)
	}
	return int(deleted), nil
}
