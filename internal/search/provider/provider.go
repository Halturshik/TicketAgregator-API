package provider

import (
	"context"
	"sort"
	"sync"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

type Request struct {
	Input  search.SearchInput
	From   search.City
	To     search.City
	Cities []search.City
	Count  int
}

type Provider interface {
	Transport() string
	Generate(ctx context.Context, req Request) ([]search.Offer, error)
}

type MockProvider struct {
	transport string
	carriers  []Carrier
}

func NewMockProvider(transport string, carriers []Carrier) *MockProvider {
	if len(carriers) == 0 {
		carriers = []Carrier{{Name: "Mock " + transport, Code: "MCK"}}
	}
	return &MockProvider{
		transport: transport,
		carriers:  carriers,
	}
}

func (p *MockProvider) Transport() string {
	return p.transport
}

func (p *MockProvider) Generate(ctx context.Context, req Request) ([]search.Offer, error) {
	if req.Count <= 0 {
		return []search.Offer{}, nil
	}

	workers := min(WorkerLimit, req.Count)
	jobs := make(chan int)
	results := make(chan search.Offer, req.Count)

	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				if ctx.Err() != nil {
					return
				}
				results <- p.generateOffer(req, index)
			}
		}()
	}

	go func() {
		defer close(jobs)
		for index := 0; index < req.Count; index++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- index:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	offers := make([]search.Offer, 0, req.Count)
	for offer := range results {
		offers = append(offers, offer)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sort.Slice(offers, func(i, j int) bool {
		return offers[i].Price < offers[j].Price
	})
	return offers, nil
}
