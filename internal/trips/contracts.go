package trips

import (
	"context"
	"time"
)

type Service interface {
	Lookup(ctx context.Context, input LookupInput) (*LookupResult, error)
}

type Repository interface {
	ByTicketNumber(ctx context.Context, ticketNumber string) ([]ScheduledTrip, error)
	ByRouteNumber(ctx context.Context, routeNumber string, from time.Time, to time.Time) ([]ScheduledTrip, error)
}
