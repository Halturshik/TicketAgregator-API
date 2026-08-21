package provider

import (
	"math"
	"math/rand"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/google/uuid"
)

func (p *MockProvider) generateOffer(req Request, index int) search.Offer {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano() + int64(index+1)*7919))
	isInternational := req.From.CountryID != req.To.CountryID
	pricePerPassenger := p.price(rnd, isInternational)
	price := pricePerPassenger * req.Input.Passengers

	departureDate, _ := time.Parse("2006-01-02", req.Input.Date)
	hour := rnd.Intn(24)
	minute := rnd.Intn(60/DepartureMinuteStep) * DepartureMinuteStep
	departure := departureDate.
		Add(time.Duration(hour) * time.Hour).
		Add(time.Duration(minute) * time.Minute)

	segments := []search.Segment{}
	lastArrival := departure
	needTransfer := p.transport == "avia" && isInternational && rnd.Intn(100) < TransferChancePercent

	firstCarrier := p.pickCarrier(rnd)

	if needTransfer {
		hub := p.pickHub(rnd, req.From, req.To, req.Cities)
		secondCarrier := p.pickSecondCarrier(rnd, firstCarrier)

		firstArrival := departure.Add(time.Duration(TransferFirstLegMinMinutes+rnd.Intn(TransferFirstLegMaxExtra)) * time.Minute)
		secondDeparture := firstArrival.Add(time.Duration(TransferWaitMinMinutes+rnd.Intn(TransferWaitMaxExtra)) * time.Minute)
		secondArrival := secondDeparture.Add(time.Duration(TransferSecondLegMinMinutes+rnd.Intn(TransferSecondLegMaxExtra)) * time.Minute)

		segments = append(segments,
			segment(1, req.From, hub, departure, firstArrival, firstCarrier, p.transport, rnd),
			segment(2, hub, req.To, secondDeparture, secondArrival, secondCarrier, p.transport, rnd),
		)
		lastArrival = secondArrival
	} else {
		arrival := departure.Add(time.Duration(p.duration(rnd, isInternational)) * time.Minute)
		segments = append(segments, segment(1, req.From, req.To, departure, arrival, firstCarrier, p.transport, rnd))
		lastArrival = arrival
	}

	return search.Offer{
		ID:                uuid.NewString(),
		Transport:         p.transport,
		IsInternational:   isInternational,
		Price:             price,
		PricePerPassenger: pricePerPassenger,
		BonusEarn:         int(math.Floor(float64(price) * BonusRate)),
		TransferCount:     len(segments) - 1,
		DurationMinutes:   int(lastArrival.Sub(departure).Minutes()),
		Segments:          segments,
	}
}

func (p *MockProvider) price(rnd *rand.Rand, international bool) int {
	base := BasePriceByTransport[p.transport]
	if international {
		base = int(float64(base) * InternationalPriceMultiplier)
	}
	return base + rnd.Intn(base/2)
}

func (p *MockProvider) duration(rnd *rand.Rand, international bool) int {
	switch p.transport {
	case "rail":
		return RailMinMinutes + rnd.Intn(RailMaxExtra)
	case "bus":
		return BusMinMinutes + rnd.Intn(BusMaxExtra)
	default:
		if international {
			return AviaIntlMinMinutes + rnd.Intn(AviaIntlMaxExtra)
		}
		return AviaDomesticMinMinutes + rnd.Intn(AviaDomesticMaxExtra)
	}
}

func segment(order int, from search.City, to search.City, departure time.Time, arrival time.Time, carrier Carrier, transport string, rnd *rand.Rand) search.Segment {
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
		RouteNumber:   routeNumber(transport, carrier.Code, rnd),
	}
}
