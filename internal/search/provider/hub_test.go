package provider

import (
	"math/rand"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func TestPickHubReportsMissingCandidate(t *testing.T) {
	provider := &MockProvider{}
	from := search.City{ID: 1}
	to := search.City{ID: 2}

	if _, found := provider.pickHub(rand.New(rand.NewSource(1)), from, to, []search.City{from, to}); found {
		t.Fatal("pickHub() found a hub when no intermediate hub exists")
	}
}

func TestPickHubUsesAvailableIntermediateHub(t *testing.T) {
	provider := &MockProvider{}
	from := search.City{ID: 1, Latitude: 1, Longitude: 1}
	to := search.City{ID: 2, Latitude: 5, Longitude: 5}
	hub := search.City{ID: 3, IsAirHub: true, Latitude: 3, Longitude: 3}

	got, found := provider.pickHub(rand.New(rand.NewSource(1)), from, to, []search.City{from, to, hub})
	if !found || got.ID != hub.ID {
		t.Fatalf("pickHub() = (%d, %t), want (%d, true)", got.ID, found, hub.ID)
	}
}
