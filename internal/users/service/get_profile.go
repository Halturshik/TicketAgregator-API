package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

func (s *Service) GetProfile(ctx context.Context, userID int) (*users.Profile, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		logger.Error("Ошибка при получении профиля пользователя userID=%d: %v", userID, err)
		return nil, err
	}
	return profile, nil
}
