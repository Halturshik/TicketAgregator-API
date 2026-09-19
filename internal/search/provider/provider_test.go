package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/transport"
)

func TestGenerateProducesEveryRequestedOption(t *testing.T) {
	provider := NewMockProvider(transport.Rail, []Carrier{{ID: 1, Name: "RZD", Code: "RZD"}})
	request := Request{
		Input: search.SearchInput{Date: "2026-09-10", Passengers: 12},
		From:  search.City{ID: 1, CountryID: 1},
		To:    search.City{ID: 2, CountryID: 1},
		Count: 45,
	}

	options, err := provider.Generate(context.Background(), request)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(options) != request.Count {
		t.Fatalf("Generate() count = %d, want %d", len(options), request.Count)
	}
	ids := make(map[string]struct{}, len(options))
	for _, option := range options {
		if _, duplicate := ids[option.ID]; duplicate {
			t.Fatalf("duplicate trip option id %q", option.ID)
		}
		ids[option.ID] = struct{}{}
		if option.Price != option.PricePerPassenger*request.Input.Passengers {
			t.Fatalf("option price = %d, per passenger = %d", option.Price, option.PricePerPassenger)
		}
	}
}

func TestGenerateHonorsCanceledContext(t *testing.T) {
	provider := NewMockProvider(transport.Bus, []Carrier{{ID: 1, Name: "Bus", Code: "BUS"}})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.Generate(ctx, Request{Count: 45})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Generate() error = %v, want context canceled", err)
	}
}
