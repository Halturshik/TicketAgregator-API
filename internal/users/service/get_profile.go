package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

func (s *Service) GetProfile(ctx context.Context, userID int) (*users.Profile, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if errors.Is(err, users.ErrNotFound) {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return profile, nil
}
