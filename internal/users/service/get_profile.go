package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

func (s *Service) GetProfile(ctx context.Context, userID int) (*users.Profile, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if errors.Is(err, users.ErrNotFound) {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка при получении профиля пользователя userID=%d: %v", userID, err)
		return nil, err
	}
	return profile, nil
}
