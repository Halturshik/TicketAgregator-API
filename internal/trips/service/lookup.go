package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	ordernumber "github.com/Halturshik/TicketAgregator-API/internal/orders/number"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/trips"
)

func (s *Service) Lookup(ctx context.Context, input trips.LookupInput) (*trips.LookupResult, error) {
	ticketNumber := normalizeNumber(input.TicketNumber)
	routeNumber := normalizeNumber(input.RouteNumber)
	if (ticketNumber == "") == (routeNumber == "") {
		return nil, apierror.ErrInvalidRequest
	}

	result := &trips.LookupResult{Trips: []trips.ScheduledTrip{}}
	var items []trips.ScheduledTrip
	var err error
	if ticketNumber != "" {
		if strings.TrimSpace(input.Date) != "" || !ordernumber.ValidTicket(ticketNumber) {
			return nil, apierror.ErrInvalidRequest
		}
		items, err = s.repo.ByTicketNumber(ctx, ticketNumber)
		result.TicketNumber = ticketNumber
	} else {
		if !trips.ValidRouteNumber(routeNumber) {
			return nil, apierror.ErrInvalidRequest
		}
		date, parseErr := time.Parse(search.DateLayout, strings.TrimSpace(input.Date))
		if parseErr != nil {
			return nil, apierror.ErrInvalidRequest
		}
		from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		items, err = s.repo.ByRouteNumber(ctx, routeNumber, from, from.AddDate(0, 0, 1))
		result.RouteNumber = routeNumber
		result.Date = from.Format(search.DateLayout)
	}
	if errors.Is(err, trips.ErrNotFound) {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка публичного поиска рейса: %v", err)
		return nil, err
	}
	result.Trips = items
	return result, nil
}

func normalizeNumber(value string) string {
	return strings.ToUpper(strings.Join(strings.Fields(value), " "))
}
