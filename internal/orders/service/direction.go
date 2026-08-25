package service

import (
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/transport"
)

func validateDirection(option search.TripOption, direction search.Offer, passengerCount int, usedRouteNumbers map[string]struct{}) error {
	if direction.ID == "" || direction.Transport != option.Transport || !transport.IsSupported(direction.Transport) || len(direction.Segments) == 0 {
		return fmt.Errorf("invalid direction metadata")
	}
	if direction.PricePerPassenger <= 0 || direction.Price != direction.PricePerPassenger*passengerCount {
		return fmt.Errorf(
			"invalid price: price=%d pricePerPassenger=%d passengers=%d",
			direction.Price, direction.PricePerPassenger, passengerCount,
		)
	}
	return validateSegments(direction, usedRouteNumbers)
}

func validateReturn(option search.TripOption) error {
	if option.Return == nil {
		return nil
	}
	outboundFirst := option.Outbound.Segments[0]
	outboundLast := option.Outbound.Segments[len(option.Outbound.Segments)-1]
	returnFirst := option.Return.Segments[0]
	returnLast := option.Return.Segments[len(option.Return.Segments)-1]
	if outboundFirst.FromCityID != returnLast.ToCityID || outboundLast.ToCityID != returnFirst.FromCityID ||
		option.Outbound.IsInternational != option.Return.IsInternational {
		return fmt.Errorf("invalid cached return route tripOptionID=%s", option.ID)
	}
	outboundArrival, _ := time.Parse(time.RFC3339, outboundLast.ArrivalTime)
	returnDeparture, _ := time.Parse(time.RFC3339, returnFirst.DepartureTime)
	if returnDeparture.Before(outboundArrival) {
		return fmt.Errorf("return departs before outbound arrives tripOptionID=%s", option.ID)
	}
	return nil
}
