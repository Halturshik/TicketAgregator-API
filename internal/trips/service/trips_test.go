package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/trips"
	tripservice "github.com/Halturshik/TicketAgregator-API/internal/trips/service"
)

func TestTripLookupByTicketAndRoute(t *testing.T) {
	repo := &tripRepositoryStub{items: []trips.ScheduledTrip{{RouteNumber: "SU 12345"}}}
	service := tripservice.NewService(repo)

	byTicket, err := service.Lookup(context.Background(), trips.LookupInput{TicketNumber: "av-20260821"})
	if err != nil {
		t.Fatalf("lookup by ticket: %v", err)
	}
	if repo.ticket != "AV-20260821" || byTicket.TicketNumber != repo.ticket {
		t.Fatalf("ticket number was not normalized: repo=%q result=%+v", repo.ticket, byTicket)
	}

	byRoute, err := service.Lookup(context.Background(), trips.LookupInput{
		RouteNumber: "su   12345", Date: "2026-10-10",
	})
	if err != nil {
		t.Fatalf("lookup by route: %v", err)
	}
	if repo.route != "SU 12345" || byRoute.Date != "2026-10-10" {
		t.Fatalf("route lookup was not normalized: repo=%q result=%+v", repo.route, byRoute)
	}
	if !repo.from.Equal(time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)) ||
		!repo.to.Equal(time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected UTC date interval: %s - %s", repo.from, repo.to)
	}
}

func TestTripLookupRejectsAmbiguousAndMalformedInput(t *testing.T) {
	service := tripservice.NewService(&tripRepositoryStub{})
	cases := []trips.LookupInput{
		{},
		{TicketNumber: "AV-20260821", RouteNumber: "SU 12345", Date: "2026-10-10"},
		{TicketNumber: "wrong"},
		{RouteNumber: "DROP TABLE", Date: "2026-10-10"},
		{RouteNumber: "SU 12345", Date: "10.10.2026"},
	}
	for _, input := range cases {
		if _, err := service.Lookup(context.Background(), input); !errors.Is(err, apierror.ErrInvalidRequest) {
			t.Fatalf("input %+v: expected invalid request, got %v", input, err)
		}
	}
}

type tripRepositoryStub struct {
	items    []trips.ScheduledTrip
	ticket   string
	route    string
	from, to time.Time
}

func (s *tripRepositoryStub) ByTicketNumber(_ context.Context, ticketNumber string) ([]trips.ScheduledTrip, error) {
	s.ticket = ticketNumber
	if len(s.items) == 0 {
		return nil, trips.ErrNotFound
	}
	return s.items, nil
}

func (s *tripRepositoryStub) ByRouteNumber(
	_ context.Context,
	routeNumber string,
	from time.Time,
	to time.Time,
) ([]trips.ScheduledTrip, error) {
	s.route, s.from, s.to = routeNumber, from, to
	if len(s.items) == 0 {
		return nil, trips.ErrNotFound
	}
	return s.items, nil
}
