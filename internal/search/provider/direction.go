package provider

import (
	"math/rand"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/transport"
)

func (p *MockProvider) generateDirection(
	req Request,
	from search.City,
	to search.City,
	date string,
	earliestDeparture time.Time,
	firstCarrier Carrier,
	usedRouteNumbers map[string]struct{},
	offerID string,
	rnd *rand.Rand,
) search.Offer {
	isInternational := from.CountryID != to.CountryID
	pricePerPassenger := p.price(rnd, isInternational)
	price := pricePerPassenger * req.Input.Passengers

	departureDate, _ := time.Parse(search.DateLayout, date)
	departure := randomDeparture(departureDate, earliestDeparture, rnd)
	segments, lastArrival := p.generateSegments(
		from, to, departure, isInternational, firstCarrier, usedRouteNumbers, req.Cities, rnd,
	)

	return search.Offer{
		ID:                offerID,
		Transport:         p.transport,
		IsInternational:   isInternational,
		Price:             price,
		PricePerPassenger: pricePerPassenger,
		TransferCount:     len(segments) - 1,
		DurationMinutes:   int(lastArrival.Sub(departure).Minutes()),
		Segments:          segments,
	}
}

func (p *MockProvider) generateSegments(
	from search.City,
	to search.City,
	departure time.Time,
	isInternational bool,
	firstCarrier Carrier,
	usedRouteNumbers map[string]struct{},
	cities []search.City,
	rnd *rand.Rand,
) ([]search.Segment, time.Time) {
	needTransfer := p.transport == transport.Avia &&
		isInternational && rnd.Intn(ProbabilityScale) < TransferChancePercent
	if needTransfer {
		if hub, found := p.pickHub(rnd, from, to, cities); found {
			return p.transferSegments(from, hub, to, departure, firstCarrier, usedRouteNumbers, rnd)
		}
	}

	arrival := departure.Add(time.Duration(p.duration(rnd, isInternational)) * time.Minute)
	segments := []search.Segment{
		segment(1, from, to, departure, arrival, firstCarrier, p.transport, usedRouteNumbers, rnd),
	}
	return segments, arrival
}
