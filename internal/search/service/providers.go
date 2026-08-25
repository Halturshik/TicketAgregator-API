package service

import (
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/search/provider"
	transportpkg "github.com/Halturshik/TicketAgregator-API/internal/transport"
)

func newProviders(carriers []search.CarrierConfig) (map[string]provider.Provider, error) {
	carriersByType := carriersByTransport(carriers)
	for _, transport := range []string{transportpkg.Avia, transportpkg.Rail, transportpkg.Bus} {
		if len(carriersByType[transport]) == 0 {
			return nil, fmt.Errorf("no carriers configured for transport %s", transport)
		}
	}
	items := []provider.Provider{
		provider.NewMockProvider(transportpkg.Avia, carriersByType[transportpkg.Avia]),
		provider.NewMockProvider(transportpkg.Rail, carriersByType[transportpkg.Rail]),
		provider.NewMockProvider(transportpkg.Bus, carriersByType[transportpkg.Bus]),
	}

	result := make(map[string]provider.Provider, len(items))
	for _, item := range items {
		result[item.Transport()] = item
	}
	return result, nil
}

func carriersByTransport(items []search.CarrierConfig) map[string][]provider.Carrier {
	result := map[string][]provider.Carrier{
		transportpkg.Avia: {},
		transportpkg.Rail: {},
		transportpkg.Bus:  {},
	}
	for _, carrier := range items {
		if _, supported := result[carrier.TransportType]; !supported {
			continue
		}
		result[carrier.TransportType] = append(result[carrier.TransportType], provider.Carrier{
			ID: carrier.ID, Name: carrier.Name, Code: carrier.Code,
		})
	}
	return result
}
