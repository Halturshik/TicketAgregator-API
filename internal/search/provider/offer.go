package provider

import (
	"math/rand"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (p *MockProvider) generateTripOption(req Request, index int) search.TripOption {
	seed := req.Seed + int64(index+1)*RandomSeedStride
	rnd := rand.New(rand.NewSource(seed))
	return p.generateTripOptionWithRand(req, index, rnd)
}

func (p *MockProvider) generateTripOptionWithRand(req Request, index int, rnd *rand.Rand) search.TripOption {
	usedRouteNumbers := make(map[string]struct{})
	firstCarrier := p.pickCarrier(rnd)
	outbound := p.generateDirection(
		req, req.From, req.To, req.Input.Date, time.Time{}, firstCarrier,
		usedRouteNumbers, stableID(req.Seed, index, "outbound"), rnd,
	)
	scheduleID := stableID(req.Seed, index, "schedule")

	option := search.TripOption{
		ID:                scheduleID,
		ScheduleID:        scheduleID,
		Transport:         p.transport,
		Price:             outbound.Price,
		PricePerPassenger: outbound.PricePerPassenger,
		Outbound:          outbound,
	}
	if req.Input.ReturnDate != "" {
		returnCarrier := firstCarrier
		if rnd.Intn(ProbabilityScale) >= SameCarrierOnReturnPercent {
			returnCarrier = p.pickDifferentCarrier(rnd, firstCarrier)
		}
		outboundArrival, _ := time.Parse(time.RFC3339, outbound.Segments[len(outbound.Segments)-1].ArrivalTime)
		returnOffer := p.generateDirection(
			req, req.To, req.From, req.Input.ReturnDate, outboundArrival.Add(ReturnConnectionTime),
			returnCarrier, usedRouteNumbers, stableID(req.Seed, index, "return"), rnd,
		)
		option.Return = &returnOffer
		option.Price += returnOffer.Price
		option.PricePerPassenger += returnOffer.PricePerPassenger
	}
	return option
}
