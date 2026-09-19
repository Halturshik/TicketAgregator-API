package provider

import (
	"math/rand"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func TestGenerateTripOptionBuildsIndependentDirections(t *testing.T) {
	p := NewMockProvider("avia", []Carrier{
		{ID: 1, Name: "Aeroflot", Code: "SU"},
		{ID: 2, Name: "S7", Code: "S7"},
	})
	req := roundTripRequest()
	option := p.generateTripOptionWithRand(req, 0, rand.New(rand.NewSource(42)))

	if option.Return == nil {
		t.Fatal("return offer is nil")
	}
	if option.Outbound.ID == option.Return.ID {
		t.Fatal("outbound and return must be separate offers")
	}
	if option.Price != option.Outbound.Price+option.Return.Price {
		t.Fatalf("trip price = %d, want directions sum", option.Price)
	}
	if option.PricePerPassenger != option.Outbound.PricePerPassenger+option.Return.PricePerPassenger {
		t.Fatalf("price per passenger = %d, want directions sum", option.PricePerPassenger)
	}
	assertDirection(t, option.Outbound, req.From.ID, req.To.ID)
	assertDirection(t, *option.Return, req.To.ID, req.From.ID)
	outboundArrival, _ := time.Parse(time.RFC3339, option.Outbound.Segments[len(option.Outbound.Segments)-1].ArrivalTime)
	returnDeparture, _ := time.Parse(time.RFC3339, option.Return.Segments[0].DepartureTime)
	if returnDeparture.Before(outboundArrival.Add(time.Hour)) {
		t.Fatalf("return departs before trip can continue: outbound=%s return=%s", outboundArrival, returnDeparture)
	}

	seenRoutes := map[string]struct{}{}
	for _, offer := range []search.Offer{option.Outbound, *option.Return} {
		for _, segment := range offer.Segments {
			if _, exists := seenRoutes[segment.RouteNumber]; exists {
				t.Fatalf("duplicate route number across trip: %s", segment.RouteNumber)
			}
			seenRoutes[segment.RouteNumber] = struct{}{}
		}
	}
}

func TestReturnUsesSamePrimaryCarrierAboutSeventyFivePercent(t *testing.T) {
	p := NewMockProvider("avia", []Carrier{
		{ID: 1, Name: "Aeroflot", Code: "SU"},
		{ID: 2, Name: "S7", Code: "S7"},
		{ID: 3, Name: "Pobeda", Code: "DP"},
	})
	req := roundTripRequest()
	const samples = 5000
	same := 0
	for seed := int64(1); seed <= samples; seed++ {
		option := p.generateTripOptionWithRand(req, int(seed), rand.New(rand.NewSource(seed)))
		if option.Outbound.Segments[0].CarrierCode == option.Return.Segments[0].CarrierCode {
			same++
		}
	}
	percentage := float64(same) * 100 / samples
	if percentage < 72 || percentage > 78 {
		t.Fatalf("same carrier percentage = %.2f, want close to 75", percentage)
	}
}

func TestOneWayOptionHasNoReturn(t *testing.T) {
	p := NewMockProvider("rail", []Carrier{{ID: 1, Name: "RZD", Code: "RZD"}})
	req := roundTripRequest()
	req.Input.ReturnDate = ""
	option := p.generateTripOptionWithRand(req, 0, rand.New(rand.NewSource(7)))
	if option.Return != nil {
		t.Fatal("one-way option unexpectedly has return")
	}
	if option.Price != option.Outbound.Price {
		t.Fatalf("one-way price = %d, want %d", option.Price, option.Outbound.Price)
	}
}

func roundTripRequest() Request {
	from := search.City{ID: 1, Name: "Moscow", CountryID: 1}
	to := search.City{ID: 2, Name: "Paris", CountryID: 2}
	hub := search.City{ID: 3, Name: "Istanbul", CountryID: 3, IsAirHub: true}
	return Request{
		Input: search.SearchInput{Date: "2026-09-10", ReturnDate: "2026-09-20", Passengers: 2},
		From:  from, To: to, Cities: []search.City{from, to, hub}, Count: 1,
	}
}

func assertDirection(t *testing.T, offer search.Offer, fromID int, toID int) {
	t.Helper()
	if len(offer.Segments) == 0 {
		t.Fatal("offer has no segments")
	}
	if offer.Segments[0].FromCityID != fromID {
		t.Fatalf("first segment from = %d, want %d", offer.Segments[0].FromCityID, fromID)
	}
	if offer.Segments[len(offer.Segments)-1].ToCityID != toID {
		t.Fatalf("last segment to = %d, want %d", offer.Segments[len(offer.Segments)-1].ToCityID, toID)
	}
}
