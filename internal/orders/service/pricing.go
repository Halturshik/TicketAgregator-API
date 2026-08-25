package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

type orderPricing struct {
	spent   int
	earned  int
	payable int
}

func (s *Service) calculatePricing(ctx context.Context, userID *int, requestedBonus int, total int) (*orderPricing, error) {
	spent, err := s.bonusSpent(ctx, userID, requestedBonus, total)
	if err != nil {
		return nil, err
	}
	earned := 0
	if userID != nil {
		earned = bonus.Earned(total)
	}
	return &orderPricing{spent: spent, earned: earned, payable: total - spent}, nil
}

func (s *Service) bonusSpent(ctx context.Context, userID *int, requested int, total int) (int, error) {
	if userID == nil {
		if requested > 0 {
			return 0, apierror.ErrUnauthorized
		}
		return 0, nil
	}
	if requested == 0 {
		return 0, nil
	}
	requested = min(requested, bonus.MaxSpend(total))
	balance, err := s.bonus.GetBalance(ctx, *userID)
	if err != nil {
		logger.Error("Ошибка получения баланса бонусов userID=%d: %v", *userID, err)
		return 0, err
	}
	return min(requested, balance), nil
}
