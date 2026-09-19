package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	orderservice "github.com/Halturshik/TicketAgregator-API/internal/orders/service"
)

func TestOrderCleanupOrchestratesRepositoryCommands(t *testing.T) {
	repo := &cleanupRepositoryStub{deletedOrders: 3, deletedTrips: 2}
	service := orderservice.NewService(repo, nil, nil, nil, nil)

	deleted, err := service.CleanupExpired(context.Background(), 0)
	if err != nil {
		t.Fatalf("cleanup expired: %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected deleted order count 3, got %d", deleted)
	}
	if repo.orderCalls != 1 || repo.tripCalls != 1 {
		t.Fatalf("unexpected cleanup calls: orders=%d trips=%d", repo.orderCalls, repo.tripCalls)
	}
	if repo.orderLimit != orders.DefaultCleanupLimit || repo.tripLimit != orders.DefaultCleanupLimit {
		t.Fatalf("cleanup limit was not normalized: orders=%d trips=%d", repo.orderLimit, repo.tripLimit)
	}
	if !repo.orderBefore.Equal(repo.tripBefore) {
		t.Fatalf("cleanup commands used different cutoffs: orders=%v trips=%v", repo.orderBefore, repo.tripBefore)
	}
}

func TestOrderCleanupStopsAfterOrderFailure(t *testing.T) {
	expected := errors.New("delete orders")
	repo := &cleanupRepositoryStub{orderErr: expected}
	service := orderservice.NewService(repo, nil, nil, nil, nil)

	deleted, err := service.CleanupExpired(context.Background(), 100)
	if !errors.Is(err, expected) || deleted != 0 {
		t.Fatalf("unexpected cleanup result: deleted=%d err=%v", deleted, err)
	}
	if repo.tripCalls != 0 {
		t.Fatalf("orphan cleanup ran after order cleanup failure")
	}
}

func TestOrderCleanupReturnsDeletedOrdersWhenOrphanCleanupFails(t *testing.T) {
	expected := errors.New("delete orphan trips")
	repo := &cleanupRepositoryStub{deletedOrders: 2, tripErr: expected}
	service := orderservice.NewService(repo, nil, nil, nil, nil)

	deleted, err := service.CleanupExpired(context.Background(), 100)
	if !errors.Is(err, expected) || deleted != 2 {
		t.Fatalf("unexpected cleanup result: deleted=%d err=%v", deleted, err)
	}
}

type cleanupRepositoryStub struct {
	deletedOrders int
	deletedTrips  int
	orderErr      error
	tripErr       error
	orderCalls    int
	tripCalls     int
	orderLimit    int
	tripLimit     int
	orderBefore   time.Time
	tripBefore    time.Time
}

func (s *cleanupRepositoryStub) Create(context.Context, orders.CreateOrderParams) (*orders.Order, error) {
	return nil, nil
}

func (s *cleanupRepositoryStub) List(context.Context, int, orders.ListFilter) (*orders.HistoryPage, error) {
	return nil, nil
}

func (s *cleanupRepositoryStub) DeleteExpiredOrders(
	_ context.Context,
	before time.Time,
	limit int,
) (int, error) {
	s.orderCalls++
	s.orderBefore = before
	s.orderLimit = limit
	return s.deletedOrders, s.orderErr
}

func (s *cleanupRepositoryStub) DeleteOrphanTrips(
	_ context.Context,
	before time.Time,
	limit int,
) (int, error) {
	s.tripCalls++
	s.tripBefore = before
	s.tripLimit = limit
	return s.deletedTrips, s.tripErr
}
