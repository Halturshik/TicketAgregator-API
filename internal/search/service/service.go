package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/search/provider"
	"github.com/Halturshik/TicketAgregator-API/internal/search/repository"
	searchstore "github.com/Halturshik/TicketAgregator-API/internal/search/store"
)

type Service struct {
	repo      *repository.Repository
	store     *searchstore.Store
	providers map[string]provider.Provider
}

func NewService(repo *repository.Repository, store *searchstore.Store) search.Service {
	carriersByType := loadCarriersByType(repo)

	providers := []provider.Provider{
		provider.NewMockProvider("avia", carriersByType["avia"]),
		provider.NewMockProvider("rail", carriersByType["rail"]),
		provider.NewMockProvider("bus", carriersByType["bus"]),
	}
	byTransport := make(map[string]provider.Provider, len(providers))
	for _, item := range providers {
		byTransport[item.Transport()] = item
	}

	return &Service{repo: repo, store: store, providers: byTransport}
}

func loadCarriersByType(repo *repository.Repository) map[string][]provider.Carrier {
	result := map[string][]provider.Carrier{
		"avia": {},
		"rail": {},
		"bus":  {},
	}

	items, err := repo.ListCarriers(context.Background())
	if err != nil {
		logger.Error("Не удалось загрузить carriers, будут fallback-заглушки: %v", err)
		return result
	}

	for _, c := range items {
		result[c.TransportType] = append(result[c.TransportType], provider.Carrier{
			ID: c.ID, Name: c.Name, Code: c.Code,
		})
	}
	return result
}
