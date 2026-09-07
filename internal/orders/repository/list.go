package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (r *Repository) List(ctx context.Context, userID int, filter orders.ListFilter) (*orders.HistoryPage, error) {
	total, err := r.countHistoryOrders(ctx, userID, filter.Transport)
	if err != nil {
		return nil, err
	}
	items, err := r.listHistoryOrders(ctx, userID, filter)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		if err := r.loadHistoryTickets(ctx, items); err != nil {
			return nil, fmt.Errorf("load order history tickets: %w", err)
		}
	}
	return &orders.HistoryPage{
		Total: total, Offset: filter.Offset, Limit: filter.Limit, Items: items,
	}, nil
}
