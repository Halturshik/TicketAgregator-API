package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (r *Repository) loadBooking(ctx context.Context, orderID int) (*bookingaccess.Booking, error) {
	var booking bookingaccess.Booking
	var userID sql.NullInt64
	err := r.db.QueryRowContext(ctx, `
		SELECT o.id, o.order_number, o.user_id, COALESCE(o.guest_email, ''),
			COALESCE(o.guest_payment_token::text, ''), o.status, o.total_price,
			o.current_total_price, o.bonus_spent, o.bonus_earned, o.created_at,
			(SELECT COUNT(*) FROM order_passengers op WHERE op.order_id = o.id)
		FROM orders o
		WHERE o.id = $1 AND o.status IN ($2, $3, $4)`,
		orderID,
		orders.OrderStatusPaid,
		orders.OrderStatusPartiallyRefunded,
		orders.OrderStatusRefunded,
	).Scan(
		&booking.ID,
		&booking.OrderNumber,
		&userID,
		&booking.GuestEmail,
		&booking.GuestPaymentToken,
		&booking.Status,
		&booking.TotalPrice,
		&booking.CurrentTotalPrice,
		&booking.BonusSpent,
		&booking.BonusEarned,
		&booking.CreatedAt,
		&booking.PassengerCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, bookingaccess.ErrNotFound
		}
		return nil, fmt.Errorf("load booking: %w", err)
	}
	if userID.Valid {
		value := int(userID.Int64)
		booking.UserID = &value
	}
	if err := r.loadTickets(ctx, &booking); err != nil {
		return nil, err
	}
	return &booking, nil
}
