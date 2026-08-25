package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
)

func (r *Repository) List(ctx context.Context, userID int, limit int, offset int) ([]bonus.Transaction, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, order_id, type, amount, created_at
		FROM bonus_transactions
		WHERE user_id = $1
		ORDER BY id DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list bonus transactions: %w", err)
	}
	defer rows.Close()

	result := make([]bonus.Transaction, 0)
	for rows.Next() {
		var item bonus.Transaction
		var orderID sql.NullInt64
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &orderID, &item.Type, &item.Amount, &createdAt); err != nil {
			return nil, fmt.Errorf("scan bonus transaction: %w", err)
		}
		if orderID.Valid {
			value := int(orderID.Int64)
			item.OrderID = &value
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bonus transactions: %w", err)
	}
	return result, nil
}
