package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (t *transaction) LockForPayment(ctx context.Context, orderID int) (orders.PaymentData, error) {
	var data orders.PaymentData
	var userID sql.NullInt64
	var guestToken sql.NullString
	err := t.tx.QueryRowContext(ctx, `
		SELECT id, user_id, guest_payment_token, status, payable_amount, bonus_spent, bonus_earned, expires_at
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`, orderID).Scan(
		&data.ID, &userID, &guestToken, &data.Status,
		&data.PayableAmount, &data.BonusSpent, &data.BonusEarned, &data.ExpiresAt,
	)
	if userID.Valid {
		value := int(userID.Int64)
		data.UserID = &value
	}
	if guestToken.Valid {
		data.GuestToken = guestToken.String
	}
	if err == sql.ErrNoRows {
		return data, orders.ErrNotFound
	}
	return data, err
}

func (t *transaction) MarkExpired(ctx context.Context, orderID int) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE orders SET status = $2 WHERE id = $1`, orderID, orders.OrderStatusExpired)
	if err != nil {
		return fmt.Errorf("mark order expired: %w", err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return fmt.Errorf("mark order expired: %w", err)
	}
	result, err = t.tx.ExecContext(ctx, `UPDATE tickets SET status = $2 WHERE order_id = $1`, orderID, orders.TicketStatusExpired)
	if err != nil {
		return fmt.Errorf("mark tickets expired: %w", err)
	}
	if err := requireAtLeastOneRow(result); err != nil {
		return fmt.Errorf("mark tickets expired: %w", err)
	}
	return nil
}

func (t *transaction) MarkPaid(ctx context.Context, orderID int) error {
	result, err := t.tx.ExecContext(ctx, `UPDATE orders SET status = $2, paid_at = NOW() WHERE id = $1`, orderID, orders.OrderStatusPaid)
	if err != nil {
		return fmt.Errorf("mark order paid: %w", err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return fmt.Errorf("mark order paid: %w", err)
	}
	result, err = t.tx.ExecContext(ctx, `UPDATE tickets SET status = $2 WHERE order_id = $1`, orderID, orders.TicketStatusPaid)
	if err != nil {
		return fmt.Errorf("mark tickets paid: %w", err)
	}
	if err := requireAtLeastOneRow(result); err != nil {
		return fmt.Errorf("mark tickets paid: %w", err)
	}
	return nil
}
