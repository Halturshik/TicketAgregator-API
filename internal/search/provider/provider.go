package provider

import (
	"context"
	"sync"

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

	workers := min(WorkerLimit, req.Count)
	jobs := make(chan int)
	results := make(chan search.TripOption, req.Count)

	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				if ctx.Err() != nil {
					return
				}
				results <- p.generateTripOption(req, index)
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

	options := make([]search.TripOption, 0, req.Count)
	for option := range results {
		options = append(options, option)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return options, nil
}
