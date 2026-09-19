package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

func TestGenerateTripOptionsCallsAllSuppliersConcurrently(t *testing.T) {
	gateway := &concurrentSupplierGateway{release: make(chan struct{})}
	service := &Service{supplier: gateway, providerCodes: supplier.DefaultProviders}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	items, err := service.generateTripOptions(ctx, "avia", search.SearchInput{Passengers: 1}, searchRoute{}, 3, 1)
	if err != nil {
		t.Fatalf("generateTripOptions() error = %v", err)
	}
	if gateway.maxActive.Load() != int32(len(supplier.DefaultProviders)) {
		t.Fatalf("maximum concurrent suppliers = %d, want %d", gateway.maxActive.Load(), len(supplier.DefaultProviders))
	}
	if len(items) != 3 {
		t.Fatalf("generated items = %d, want 3", len(items))
	}
}

func TestMergeOffersKeepsCheapestSupplierPerScheduleAndFare(t *testing.T) {
	policy, _ := fare.Policy(fare.Standard)
	items := []search.TripOption{
		{ID: "expensive", ScheduleID: "schedule", SupplierCode: "vertex", FareType: fare.Standard, RefundPolicy: policy, Price: 1200},
		{ID: "cheap", ScheduleID: "schedule", SupplierCode: "atlas", FareType: fare.Standard, RefundPolicy: policy, Price: 1000},
		{ID: "flexible", ScheduleID: "schedule", SupplierCode: "nexus", FareType: fare.Flexible, Price: 1400},
	}
	merged := mergeOffers(items)
	if len(merged) != 2 {
		t.Fatalf("merged items = %d, want 2", len(merged))
	}
	for _, item := range merged {
		if item.FareType == fare.Standard && item.ID != "cheap" {
			t.Fatalf("standard offer = %q, want cheapest", item.ID)
		}
	}
}

func TestGenerateTripOptionsFailsOnlyWhenEverySupplierFails(t *testing.T) {
	service := &Service{
		supplier:      failingSupplierGateway{},
		providerCodes: supplier.DefaultProviders,
	}
	_, err := service.generateTripOptions(context.Background(), "avia", search.SearchInput{Passengers: 1}, searchRoute{}, 3, 1)
	if err == nil {
		t.Fatal("generateTripOptions() expected all-suppliers error")
	}
}

type concurrentSupplierGateway struct {
	started   atomic.Int32
	active    atomic.Int32
	maxActive atomic.Int32
	release   chan struct{}
	once      sync.Once
}

func (g *concurrentSupplierGateway) SearchOffers(ctx context.Context, request supplier.SearchRequest) ([]supplier.TripOption, error) {
	active := g.active.Add(1)
	defer g.active.Add(-1)
	for {
		maximum := g.maxActive.Load()
		if active <= maximum || g.maxActive.CompareAndSwap(maximum, active) {
			break
		}
	}
	if g.started.Add(1) == int32(len(supplier.DefaultProviders)) {
		g.once.Do(func() { close(g.release) })
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-g.release:
	}
	return []supplier.TripOption{{
		ID: request.ProviderCode, ScheduleID: request.ProviderCode,
		SupplierCode: request.ProviderCode, FareType: fare.Standard, Price: 1000,
	}}, nil
}

func (g *concurrentSupplierGateway) QuoteRefund(context.Context, supplier.RefundQuoteRequest) (*supplier.RefundQuote, error) {
	return nil, errors.New("not implemented")
}

func (g *concurrentSupplierGateway) ExecuteRefund(context.Context, supplier.ExecuteRefundRequest) (*supplier.ExecuteRefundResult, error) {
	return nil, errors.New("not implemented")
}

type failingSupplierGateway struct{}

func (f failingSupplierGateway) SearchOffers(context.Context, supplier.SearchRequest) ([]supplier.TripOption, error) {
	return nil, fmt.Errorf("supplier failed")
}

func (f failingSupplierGateway) QuoteRefund(context.Context, supplier.RefundQuoteRequest) (*supplier.RefundQuote, error) {
	return nil, fmt.Errorf("supplier failed")
}

func (f failingSupplierGateway) ExecuteRefund(context.Context, supplier.ExecuteRefundRequest) (*supplier.ExecuteRefundResult, error) {
	return nil, fmt.Errorf("supplier failed")
}
