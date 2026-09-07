package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (r *Repository) countHistoryOrders(ctx context.Context, userID int, transport string) (int, error) {
	where, args := historyOrderFilter(userID, transport)
	var total int
	if err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders o"+where, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count order history: %w", err)
	}
	return total, nil
}

func (r *Repository) listHistoryOrders(ctx context.Context, userID int, filter orders.ListFilter) ([]orders.HistoryOrder, error) {
	where, args := historyOrderFilter(userID, filter.Transport)
	args = append(args, filter.Limit, filter.Offset)
	query := `
		SELECT o.id, o.order_number, o.status, o.current_total_price,
			o.bonus_spent, o.bonus_earned, o.created_at
		FROM orders o
	` + where + fmt.Sprintf(" ORDER BY o.created_at DESC, o.id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list order history: %w", err)
	}
	defer rows.Close()

	items := make([]orders.HistoryOrder, 0, filter.Limit)
	for rows.Next() {
		var item orders.HistoryOrder
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID, &item.OrderNumber, &item.Status, &item.CurrentTotalPrice,
			&item.BonusSpent, &item.BonusEarned, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan order history: %w", err)
		}
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		item.Directions = []orders.HistoryDirection{}
		item.Tickets = []orders.HistoryTicket{}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order history: %w", err)
	}
	return items, nil
}

func historyOrderFilter(userID int, transport string) (string, []any) {
	where := " WHERE o.user_id = $1 AND o.status IN ($2, $3, $4)"
	args := []any{
		userID,
		orders.OrderStatusPaid,
		orders.OrderStatusPartiallyRefunded,
		orders.OrderStatusRefunded,
	}
	if transport != "" {
		args = append(args, transport)
		where += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM tickets filtered_ticket
			WHERE filtered_ticket.order_id = o.id AND filtered_ticket.transport_type = $%d
		)`, len(args))
	}
	return where, args
}
