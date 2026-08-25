package provider

import (
	"math/rand"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (p *MockProvider) transferSegments(
	from search.City,
	hub search.City,
	to search.City,
	departure time.Time,
	firstCarrier Carrier,
	usedRouteNumbers map[string]struct{},
	rnd *rand.Rand,
) ([]search.Segment, time.Time) {
	secondCarrier := p.pickSecondCarrier(rnd, firstCarrier)
	firstArrival := departure.Add(randomMinutes(rnd, TransferFirstLegMinMinutes, TransferFirstLegMaxExtra))
	secondDeparture := firstArrival.Add(randomMinutes(rnd, TransferWaitMinMinutes, TransferWaitMaxExtra))
	secondArrival := secondDeparture.Add(randomMinutes(rnd, TransferSecondLegMinMinutes, TransferSecondLegMaxExtra))

	segments := []search.Segment{
		segment(1, from, hub, departure, firstArrival, firstCarrier, p.transport, usedRouteNumbers, rnd),
		segment(2, hub, to, secondDeparture, secondArrival, secondCarrier, p.transport, usedRouteNumbers, rnd),
	}
	return segments, secondArrival
}

func randomMinutes(rnd *rand.Rand, minimum int, maximumExtra int) time.Duration {
	return time.Duration(minimum+rnd.Intn(maximumExtra)) * time.Minute
}
