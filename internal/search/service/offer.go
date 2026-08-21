package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/redis/go-redis/v9"
)

func (s *Service) GetOffer(ctx context.Context, searchID string, offerID string) (*search.Offer, error) {
	result, err := s.GetCachedResult(ctx, searchID)
	if err != nil {
		return nil, err
	}
	for _, offer := range result.Items {
		if offer.ID == offerID {
			return &offer, nil
		}
	}
	return nil, apierror.ErrNotFound
}

func (s *Service) GetCachedResult(ctx context.Context, searchID string) (*search.CachedResult, error) {
	result, err := s.store.Get(ctx, searchID)
	if err == redis.Nil {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}
