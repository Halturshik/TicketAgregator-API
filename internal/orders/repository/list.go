package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (r *Repository) List(ctx context.Context, userID int, filter orders.ListFilter) ([]orders.Order, error) {
	query := `
		SELECT DISTINCT o.id, o.status, o.total_price, o.bonus_spent, o.bonus_earned, o.payable_amount
		FROM orders o
		LEFT JOIN tickets t ON t.order_id = o.id
		WHERE o.user_id = $1
	`
	args := []any{userID}
	if filter.Transport != "" {
		args = append(args, filter.Transport)
		query += fmt.Sprintf(" AND t.transport_type = $%d", len(args))
	}
	args = append(args, filter.Limit, filter.Offset)
	query += fmt.Sprintf(" ORDER BY o.id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	result := []orders.Order{}
	for rows.Next() {
		var order orders.Order
		id := userID
		order.UserID = &id
		if err := rows.Scan(&order.ID, &order.Status, &order.TotalPrice, &order.BonusSpent, &order.BonusEarned, &order.PayableAmount); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		result = append(result, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}
	return result, nil
}
