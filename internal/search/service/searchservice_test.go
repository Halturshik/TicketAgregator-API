package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/Halturshik/TicketAgregator-API/internal/transport"
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

func TestNormalizeInputValidatesDateAndPassengerRange(t *testing.T) {
	now := time.Date(2026, time.August, 22, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		date       string
		passengers int
		wantError  bool
	}{
		{name: "today", date: "2026-08-22", passengers: 1},
		{name: "twelve months", date: "2027-08-22", passengers: 12},
		{name: "past date", date: "2026-08-21", passengers: 1, wantError: true},
		{name: "more than twelve months", date: "2027-08-23", passengers: 1, wantError: true},
		{name: "zero passengers", date: "2026-08-22", passengers: 0, wantError: true},
		{name: "too many passengers", date: "2026-08-22", passengers: 13, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := search.SearchInput{Date: tt.date, Passengers: tt.passengers}
			_, err := normalizeInput(&input, now)
			if (err != nil) != tt.wantError {
				t.Fatalf("normalizeInput() error = %v, wantError = %t", err, tt.wantError)
			}
		})
	}
}

func TestNormalizeInputValidatesReturnDate(t *testing.T) {
	now := time.Date(2026, time.August, 22, 15, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		returnDate string
		wantError  bool
	}{
		{name: "one way"},
		{name: "same day", returnDate: "2026-09-10", wantError: true},
		{name: "after outbound", returnDate: "2026-09-20"},
		{name: "before outbound", returnDate: "2026-09-09", wantError: true},
		{name: "outside horizon", returnDate: "2027-08-23", wantError: true},
		{name: "invalid format", returnDate: "20.09.2026", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := search.SearchInput{Date: "2026-09-10", ReturnDate: tt.returnDate, Passengers: 1}
			_, err := normalizeInput(&input, now)
			if (err != nil) != tt.wantError {
				t.Fatalf("normalizeInput() error = %v, wantError = %t", err, tt.wantError)
			}
		})
	}
}

func TestGenerationRangeDependsOnSearchHorizon(t *testing.T) {
	today := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		date     time.Time
		base     int
		maxExtra int
	}{
		{name: "first three months", date: today.AddDate(0, 3, 0), base: 52, maxExtra: 36},
		{name: "four to six months", date: today.AddDate(0, 4, 0), base: 31, maxExtra: 14},
		{name: "seven to nine months", date: today.AddDate(0, 7, 0), base: 19, maxExtra: 9},
		{name: "ten to twelve months", date: today.AddDate(0, 10, 0), base: 6, maxExtra: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := generationRange(tt.date, today)
			if rule.base != tt.base || rule.maxExtra != tt.maxExtra {
				t.Fatalf("generationRange() = %+v, want base=%d maxExtra=%d", rule, tt.base, tt.maxExtra)
			}
			if got := generatedOffersCount(tt.date, today, int64(tt.maxExtra)); got != tt.base+tt.maxExtra {
				t.Fatalf("generatedOffersCount() = %d, want %d", got, tt.base+tt.maxExtra)
			}
		})
	}
}

func TestPageHandlesMaximumOffsetWithoutIntegerOverflow(t *testing.T) {
	service := &Service{}
	result := &search.CachedResult{
		SearchID: "search-1",
		Total:    1,
		Items:    []search.TripOption{{ID: "trip-1"}},
	}

	page := service.page(result, int(^uint(0)>>1), DefaultPageSize, nil)
	if page.Offset != len(result.Items) || len(page.Items) != 0 {
		t.Fatalf("page = offset:%d items:%d, want offset:%d items:0", page.Offset, len(page.Items), len(result.Items))
	}
}

func TestLoadRouteMapsWrappedCityNotFound(t *testing.T) {
	svc := &Service{repo: &fakeSearchRepository{cityErr: fmt.Errorf("get city: %w", search.ErrCityNotFound)}}
	_, err := svc.loadRoute(context.Background(), search.SearchInput{FromCityID: 999, ToCityID: 1})
	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("loadRoute() error = %v, want not found", err)
	}
}

func TestNewServiceRequiresCarrierForEveryTransport(t *testing.T) {
	_, err := NewService(&fakeSearchRepository{}, &fakeSearchStore{}, &fakeSupplierGateway{}, []string{supplier.ProviderAtlas}, []search.CarrierConfig{
		{TransportType: transport.Avia, Name: "Avia", Code: "AV"},
	})
	if err == nil {
		t.Fatal("NewService() expected missing carrier configuration error")
	}

	_, err = NewService(&fakeSearchRepository{}, &fakeSearchStore{}, &fakeSupplierGateway{}, []string{supplier.ProviderAtlas}, []search.CarrierConfig{
		{TransportType: transport.Avia, Name: "Avia", Code: "AV"},
		{TransportType: transport.Rail, Name: "Rail", Code: "RL"},
		{TransportType: transport.Bus, Name: "Bus", Code: "BS"},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
}

func TestGetCachedResultMapsWrappedStoreNotFound(t *testing.T) {
	svc := &Service{store: &fakeSearchStore{getErr: fmt.Errorf("redis: %w", search.ErrCachedResultNotFound)}}
	_, err := svc.GetCachedResult(context.Background(), "missing")
	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("GetCachedResult() error = %v, want not found", err)
	}
}

type fakeSearchRepository struct {
	city    *search.City
	cityErr error
}

func (f *fakeSearchRepository) GetCity(context.Context, int) (*search.City, error) {
	return f.city, f.cityErr
}
func (f *fakeSearchRepository) ListCities(context.Context) ([]search.City, error) {
	return nil, nil
}
func (f *fakeSearchRepository) ListCitiesPublic(context.Context) ([]search.CityPublic, error) {
	return nil, nil
}
func (f *fakeSearchRepository) ListCarriers(context.Context) ([]search.CarrierConfig, error) {
	return nil, nil
}

type fakeSearchStore struct {
	getErr error
}

type fakeSupplierGateway struct{}

func (f *fakeSupplierGateway) SearchOffers(context.Context, supplier.SearchRequest) ([]supplier.TripOption, error) {
	return nil, nil
}

func (f *fakeSupplierGateway) QuoteRefund(context.Context, supplier.RefundQuoteRequest) (*supplier.RefundQuote, error) {
	return nil, nil
}

func (f *fakeSupplierGateway) ExecuteRefund(context.Context, supplier.ExecuteRefundRequest) (*supplier.ExecuteRefundResult, error) {
	return nil, nil
}

func (f *fakeSearchStore) Save(context.Context, *search.CachedResult) error { return nil }
func (f *fakeSearchStore) Get(context.Context, string) (*search.CachedResult, error) {
	return nil, f.getErr
}
