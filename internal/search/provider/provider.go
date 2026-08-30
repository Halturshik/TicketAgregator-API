package provider

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

type MockProvider struct {
	transport string
	carriers  []Carrier
}

func NewMockProvider(transport string, carriers []Carrier) *MockProvider {
	if len(carriers) == 0 {
		carriers = []Carrier{{Name: "Mock " + transport, Code: DefaultCarrierCode}}
	}
	return &MockProvider{
		transport: transport,
		carriers:  carriers,
	}
}

func (p *MockProvider) Transport() string {
	return p.transport
}

func (p *MockProvider) Generate(ctx context.Context, req Request) ([]search.TripOption, error) {
	if req.Count <= 0 {
		return []search.TripOption{}, nil
	}

	options := make([]search.TripOption, 0, req.Count)
	for index := 0; index < req.Count; index++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		options = append(options, p.generateTripOption(req, index))
	}
	return options, nil
}
