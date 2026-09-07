package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (r *Repository) Create(ctx context.Context, params orders.CreateOrderParams) (*orders.Order, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin order tx: %w", err)
	}
	defer tx.Rollback()

	order, err := insertOrder(ctx, tx, params)
	if err != nil {
		return nil, err
	}
	orderPassengerIDs, err := insertOrderPassengers(ctx, tx, order.ID, params.Passengers)
	if err != nil {
		return nil, err
	}
	order.Tickets, err = insertTickets(ctx, tx, order.ID, orderPassengerIDs, params)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit order tx: %w", err)
	}
	return order, nil
}

func insertOrder(ctx context.Context, tx *sql.Tx, params orders.CreateOrderParams) (*orders.Order, error) {
	var nullableUserID sql.NullInt64
	if params.UserID != nil {
		nullableUserID = sql.NullInt64{Int64: int64(*params.UserID), Valid: true}
	}

	var order orders.Order
	var expiresAt time.Time
	err := tx.QueryRowContext(ctx, `
		INSERT INTO orders
			(order_number, user_id, guest_email, guest_payment_token, total_price, current_total_price,
			 bonus_spent, bonus_earned, payable_amount, expires_at)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, '')::UUID, $5, $6, $7, $8, $9, $10)
		RETURNING id, order_number, status, total_price, current_total_price,
			bonus_spent, bonus_earned, payable_amount, expires_at
	`, params.OrderNumber, nullableUserID, params.GuestEmail, params.GuestPaymentToken,
		params.TotalPrice, params.CurrentTotalPrice, params.BonusSpent, params.BonusEarned,
		params.PayableAmount, params.ExpiresAt).
		Scan(&order.ID, &order.OrderNumber, &order.Status, &order.TotalPrice, &order.CurrentTotalPrice,
			&order.BonusSpent, &order.BonusEarned, &order.PayableAmount, &expiresAt)
	if err != nil {
		if isOrderNumberConflict(err) {
			return nil, orders.ErrOrderNumberConflict
		}
		return nil, fmt.Errorf("insert order: %w", err)
	}
	order.UserID = params.UserID
	order.GuestEmail = params.GuestEmail
	order.GuestPaymentToken = params.GuestPaymentToken
	order.ExpiresAt = expiresAt.Format(time.RFC3339)
	return &order, nil
}
