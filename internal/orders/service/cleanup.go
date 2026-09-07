package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) CleanupExpired(ctx context.Context, limit int) (int, error) {
	limit = orders.NormalizeCleanupLimit(limit)
	before := s.now().UTC().Add(-orders.ExpiredOrderRetention)
	deletedOrders, err := s.repo.DeleteExpiredOrders(ctx, before, limit)
	if err != nil {
		logger.Error("Ошибка удаления просроченных неоплаченных заказов: %v", err)
		return 0, err
	}
	if deletedOrders > 0 {
		logger.Info("Удалены просроченные неоплаченные заказы: count=%d", deletedOrders)
	}

	deletedTrips, err := s.repo.DeleteOrphanTrips(ctx, before, limit)
	if err != nil {
		logger.Error("Ошибка удаления неиспользуемых рейсов: %v", err)
		return deletedOrders, err
	}
	if deletedTrips > 0 {
		logger.Info("Удалены неиспользуемые рейсы: count=%d", deletedTrips)
	}
	return deletedOrders, nil
}
