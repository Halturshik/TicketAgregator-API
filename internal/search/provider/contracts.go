package provider

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

type Provider interface {
	Transport() string
	Generate(ctx context.Context, req Request) ([]search.TripOption, error)
}
