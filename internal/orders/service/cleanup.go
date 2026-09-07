package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) CleanupExpired(ctx context.Context, limit int) (int, error) {
	limit = orders.NormalizeCleanupLimit(limit)
	before := s.now().UTC().Add(-orders.ExpiredOrderRetention)
	deleted, err := s.repo.DeleteExpired(ctx, before, limit)
	if err != nil {
		logger.Error("Ошибка удаления просроченных неоплаченных заказов: %v", err)
		return 0, err
	}
	if deleted > 0 {
		logger.Info("Удалены просроченные неоплаченные заказы: count=%d", deleted)
	}
	return deleted, nil
}
