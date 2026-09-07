package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (r *Repository) PublicByTicket(ctx context.Context, ticketNumber string) (*bookingaccess.Booking, error) {
	return r.byLocator(ctx, "t.ticket_number = $1", ticketNumber, "")
}

func (r *Repository) PublicByOrder(ctx context.Context, orderNumber string) (*bookingaccess.Booking, error) {
	return r.byLocator(ctx, "o.order_number = $1", orderNumber, "")
}

func (r *Repository) GuestByLocator(
	ctx context.Context,
	locator bookingaccess.Locator,
	email string,
) (*bookingaccess.Booking, error) {
	if locator.TicketNumber != "" {
		return r.byLocator(ctx, "t.ticket_number = $1", locator.TicketNumber, email)
	}
	return r.byLocator(ctx, "o.order_number = $1", locator.OrderNumber, email)
}

func (r *Repository) ByOrderNumber(ctx context.Context, orderNumber string) (*bookingaccess.Booking, error) {
	return r.byLocator(ctx, "o.order_number = $1", orderNumber, "")
}

func (r *Repository) byLocator(
	ctx context.Context,
	condition string,
	value string,
	guestEmail string,
) (*bookingaccess.Booking, error) {
	query := `
		SELECT DISTINCT o.id
		FROM orders o
		JOIN tickets t ON t.order_id = o.id
		WHERE ` + condition + ` AND o.status IN ($2, $3, $4)`
	args := []any{
		value,
		orders.OrderStatusPaid,
		orders.OrderStatusPartiallyRefunded,
		orders.OrderStatusRefunded,
	}
	if guestEmail != "" {
		query += " AND o.user_id IS NULL AND o.guest_email = $5"
		args = append(args, guestEmail)
	}

	var orderID int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&orderID); err != nil {
		if err == sql.ErrNoRows {
			return nil, bookingaccess.ErrNotFound
		}
		return nil, fmt.Errorf("find booking by locator: %w", err)
	}
	return r.loadBooking(ctx, orderID)
}
