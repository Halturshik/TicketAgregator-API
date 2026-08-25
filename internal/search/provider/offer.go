package provider

import (
	"math/rand"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/google/uuid"
)

func (p *MockProvider) generateTripOption(req Request, index int) search.TripOption {
	seed := time.Now().UnixNano() + int64(index+1)*RandomSeedStride
	rnd := rand.New(rand.NewSource(seed))
	return p.generateTripOptionWithRand(req, rnd)
}

func (p *MockProvider) generateTripOptionWithRand(req Request, rnd *rand.Rand) search.TripOption {
	usedRouteNumbers := make(map[string]struct{})
	firstCarrier := p.pickCarrier(rnd)
	outbound := p.generateDirection(req, req.From, req.To, req.Input.Date, time.Time{}, firstCarrier, usedRouteNumbers, rnd)

	option := search.TripOption{
		ID:                uuid.NewString(),
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
			returnCarrier, usedRouteNumbers, rnd,
		)
		option.Return = &returnOffer
		option.Price += returnOffer.Price
		option.PricePerPassenger += returnOffer.PricePerPassenger
	}
	option.BonusEarn = bonus.Earned(option.Price)
	return option
}
