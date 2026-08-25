package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/search/provider"
	"github.com/google/uuid"
)

type Service struct {
	repo      search.Repository
	store     search.Store
	providers map[string]provider.Provider
	now       func() time.Time
	newID     func() string
}

func NewService(repo search.Repository, store search.Store, carriers []search.CarrierConfig) (search.Service, error) {
	providers, err := newProviders(carriers)
	if err != nil {
		return nil, err
	}
	return &Service{
		repo:      repo,
		store:     store,
		providers: providers,
		now:       time.Now,
		newID:     uuid.NewString,
	}, nil
}
