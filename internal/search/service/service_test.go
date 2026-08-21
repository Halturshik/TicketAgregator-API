package service

import (
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func TestGroundRouteAllowedKeepsGroundRoutesLocal(t *testing.T) {
	russia := search.City{CountryID: 1, NeighborCountryIDs: []int{2, 3}}
	belarus := search.City{CountryID: 2}
	france := search.City{CountryID: 6}

	if !groundRouteAllowed(russia, belarus) {
		t.Fatal("expected neighboring rail route to be allowed")
	}
	if groundRouteAllowed(russia, france) {
		t.Fatal("expected long-distance ground route to be rejected")
	}
}
