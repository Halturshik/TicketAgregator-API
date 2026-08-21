package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

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

	result := []bonus.Transaction{}
	for rows.Next() {
		var item bonus.Transaction
		var orderID sql.NullInt64
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &orderID, &item.Type, &item.Amount, &createdAt); err != nil {
			return nil, fmt.Errorf("scan bonus transaction: %w", err)
		}
		if orderID.Valid {
			v := int(orderID.Int64)
			item.OrderID = &v
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		result = append(result, item)
	}
	return result, rows.Err()
}
