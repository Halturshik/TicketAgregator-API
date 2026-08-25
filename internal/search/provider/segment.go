package provider

import (
	"math/rand"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func segment(
	order int,
	from search.City,
	to search.City,
	departure time.Time,
	arrival time.Time,
	carrier Carrier,
	transport string,
	usedRouteNumbers map[string]struct{},
	rnd *rand.Rand,
) search.Segment {
	return search.Segment{
		Order:         order,
		FromCityID:    from.ID,
		ToCityID:      to.ID,
		FromCity:      from.Name,
		ToCity:        to.Name,
		DepartureTime: departure.Format(time.RFC3339),
		ArrivalTime:   arrival.Format(time.RFC3339),
		CarrierID:     carrier.ID,
		Carrier:       carrier.Name,
		CarrierCode:   carrier.Code,
		RouteNumber:   uniqueRouteNumber(transport, carrier.Code, usedRouteNumbers, rnd),
	}
}

func uniqueRouteNumber(transport string, carrierCode string, used map[string]struct{}, rnd *rand.Rand) string {
	for {
		number := routeNumber(transport, carrierCode, rnd)
		if _, exists := used[number]; exists {
			continue
		}
		used[number] = struct{}{}
		return number
	}
}
