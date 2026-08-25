package service

import (
	"context"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (s *Service) GetOffer(ctx context.Context, searchID string, offerID string) (*search.Offer, error) {
	result, err := s.GetCachedResult(ctx, searchID)
	if err != nil {
		return nil, err
	}
	for _, option := range result.Items {
		if option.Outbound.ID == offerID {
			return &option.Outbound, nil
		}
		if option.Return != nil && option.Return.ID == offerID {
			return option.Return, nil
		}
	}
	return nil, apierror.ErrNotFound
}

func (s *Service) GetTripOption(ctx context.Context, searchID string, tripOptionID string) (*search.TripOption, error) {
	result, err := s.GetCachedResult(ctx, searchID)
	if err != nil {
		return nil, err
	}
	for _, option := range result.Items {
		if option.ID == tripOptionID {
			return &option, nil
		}
	}
	return nil, apierror.ErrNotFound
}

func (s *Service) GetCachedResult(ctx context.Context, searchID string) (*search.CachedResult, error) {
	result, err := s.store.Get(ctx, searchID)
	if errors.Is(err, search.ErrCachedResultNotFound) {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}
