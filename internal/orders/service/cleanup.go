package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (s *Service) CleanupExpired(ctx context.Context, limit int) (int, error) {
	limit = orders.NormalizeCleanupLimit(limit)
	before := s.now().UTC().Add(-orders.ExpiredOrderRetention)
	deletedOrders, err := s.repo.DeleteExpiredOrders(ctx, before, limit)
	if err != nil {
		return 0, fmt.Errorf("cleanup expired unpaid orders: %w", err)
	}
	if deletedOrders > 0 {
		slog.InfoContext(ctx, "Удалены просроченные неоплаченные заказы", slog.Int("count", deletedOrders))
	}

	deletedTrips, err := s.repo.DeleteOrphanTrips(ctx, before, limit)
	if err != nil {
		return deletedOrders, fmt.Errorf("cleanup orphan scheduled trips: %w", err)
	}
	if deletedTrips > 0 {
		slog.InfoContext(ctx, "Удалены неиспользуемые рейсы", slog.Int("count", deletedTrips))
	}
	return deletedOrders, nil
}
