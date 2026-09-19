package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/Halturshik/TicketAgregator-API/internal/transport"
)

func TestEverySupplierSupportsEveryFareType(t *testing.T) {
	want := []string{fare.NonRefundable, fare.Standard, fare.Flexible}
	for _, code := range supplier.DefaultProviders {
		profile, err := Profile(code)
		if err != nil {
			t.Fatalf("Profile(%q): %v", code, err)
		}
		if !reflect.DeepEqual(profile.FareTypes, want) {
			t.Fatalf("supplier %q fares = %v, want %v", code, profile.FareTypes, want)
		}
	}
}

func TestSuppliersKeepSamePhysicalSchedule(t *testing.T) {
	request := Request{
		Input: search.SearchInput{Date: "2026-10-10", ReturnDate: "2026-10-20", Passengers: 3},
		From:  search.City{ID: 1, CountryID: 1, Name: "Москва"},
		To:    search.City{ID: 2, CountryID: 1, Name: "Санкт-Петербург"},
		Cities: []search.City{
			{ID: 1, CountryID: 1, Name: "Москва"},
			{ID: 2, CountryID: 1, Name: "Санкт-Петербург"},
		},
		Count: 20,
		Seed:  987654,
	}
	carriers := []Carrier{{ID: 1, Name: "Аэрофлот", Code: "SU"}}
	atlas, err := GenerateSupplierOffers(context.Background(), supplier.ProviderAtlas, transport.Avia, carriers, request)
	if err != nil {
		t.Fatalf("atlas generation: %v", err)
	}
	nexus, err := GenerateSupplierOffers(context.Background(), supplier.ProviderNexus, transport.Avia, carriers, request)
	if err != nil {
		t.Fatalf("nexus generation: %v", err)
	}
	if len(atlas) != request.Count || len(nexus) != request.Count {
		t.Fatalf("offer counts = atlas:%d nexus:%d, want %d", len(atlas), len(nexus), request.Count)
	}

	for index := range atlas {
		if atlas[index].ScheduleID != nexus[index].ScheduleID {
			t.Fatalf("schedule %d differs: %s != %s", index, atlas[index].ScheduleID, nexus[index].ScheduleID)
		}
		if !reflect.DeepEqual(atlas[index].Outbound.Segments, nexus[index].Outbound.Segments) {
			t.Fatalf("outbound schedule %d differs between suppliers", index)
		}
		if atlas[index].Return == nil || nexus[index].Return == nil || !reflect.DeepEqual(atlas[index].Return.Segments, nexus[index].Return.Segments) {
			t.Fatalf("return schedule %d differs between suppliers", index)
		}
		if atlas[index].SupplierOfferID == nexus[index].SupplierOfferID {
			t.Fatalf("supplier offer id %d unexpectedly shared", index)
		}
	}
}
