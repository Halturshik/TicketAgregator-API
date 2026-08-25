package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/payments"
)

func (t *transaction) Create(ctx context.Context, record payments.Record) (int, error) {
	var paymentID int
	err := t.tx.QueryRowContext(ctx, `
		INSERT INTO payments (order_id, amount, status, provider)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, record.OrderID, record.Amount, record.Status, record.Provider).Scan(&paymentID)
	if err != nil {
		return 0, fmt.Errorf("insert payment: %w", err)
	}
	return paymentID, nil
}
