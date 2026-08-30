package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/search/provider"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

func (s *Service) SearchOffers(ctx context.Context, request supplier.SearchRequest) ([]supplier.TripOption, error) {
	carriers := make([]provider.Carrier, 0, len(request.Carriers))
	for _, carrier := range request.Carriers {
		if carrier.TransportType == request.Transport {
			carriers = append(carriers, provider.Carrier{ID: carrier.ID, Name: carrier.Name, Code: carrier.Code})
		}
	}
	if request.ProviderCode == "" || request.Transport == "" || request.Count <= 0 ||
		request.Input.Passengers <= 0 {
		return nil, supplier.ErrInvalidSearchRequest
	}
	if _, err := provider.Profile(request.ProviderCode); err != nil {
		return nil, supplier.ErrSupplierNotFound
	}
	if len(carriers) == 0 {
		return nil, supplier.ErrNoCarriers
	}
	return provider.GenerateSupplierOffers(ctx, request.ProviderCode, request.Transport, carriers, provider.Request{
		Input: request.Input, From: request.From, To: request.To, Cities: request.Cities,
		Count: request.Count, Seed: request.Seed,
	})
}
