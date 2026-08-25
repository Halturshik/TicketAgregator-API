package service

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

type selectedTrip struct {
	directions []search.Offer
	total      int
}

func (s *Service) loadTrip(ctx context.Context, in orders.CreateOrderInput) (*selectedTrip, error) {
	result, err := s.search.GetCachedResult(ctx, in.SearchID)
	if err != nil {
		return nil, err
	}
	if result.Input.Passengers != len(in.Passengers) {
		return nil, apierror.ErrPassengerCountMismatch
	}
	option, err := selectTripOption(result.Items, in.TripOptionID)
	if err != nil {
		return nil, err
	}
	directions, total, err := validateTripOption(in.SearchID, *option, len(in.Passengers))
	if err != nil {
		return nil, err
	}
	return &selectedTrip{directions: directions, total: total}, nil
}

func selectTripOption(all []search.TripOption, id string) (*search.TripOption, error) {
	for i := range all {
		if all[i].ID == id {
			return &all[i], nil
		}
	}
	return nil, apierror.ErrNotFound
}

func validateTripOption(searchID string, option search.TripOption, passengerCount int) ([]search.Offer, int, error) {
	directions := []search.Offer{option.Outbound}
	if option.Return != nil {
		directions = append(directions, *option.Return)
	}

	total, pricePerPassenger := 0, 0
	usedRouteNumbers := make(map[string]struct{})
	for _, direction := range directions {
		if err := validateDirection(option, direction, passengerCount, usedRouteNumbers); err != nil {
			logger.Error("Некорректное предложение в кеше: searchID=%s tripOptionID=%s offerID=%s: %v", searchID, option.ID, direction.ID, err)
			return nil, 0, fmt.Errorf("invalid cached offer %s: %w", direction.ID, err)
		}
		total += direction.Price
		pricePerPassenger += direction.PricePerPassenger
	}
	if option.Price != total || option.PricePerPassenger != pricePerPassenger || option.Transport != option.Outbound.Transport {
		logger.Error("Некорректная итоговая цена варианта в кеше: searchID=%s tripOptionID=%s", searchID, option.ID)
		return nil, 0, fmt.Errorf("invalid cached trip option totals %s", option.ID)
	}
	if err := validateReturn(option); err != nil {
		logger.Error("Некорректный обратный маршрут в кеше: searchID=%s tripOptionID=%s: %v", searchID, option.ID, err)
		return nil, 0, err
	}
	return directions, total, nil
}
