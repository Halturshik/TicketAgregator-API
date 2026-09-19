//go:build integration

package integration_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	bookingrepo "github.com/Halturshik/TicketAgregator-API/internal/bookingaccess/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	orderrepo "github.com/Halturshik/TicketAgregator-API/internal/orders/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	triprepo "github.com/Halturshik/TicketAgregator-API/internal/trips/repository"
)

func TestPostgresBookingTripLifecycle(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()

	ordersRepository := orderrepo.NewRepository(db)
	departure := time.Now().UTC().AddDate(0, 2, 0).Truncate(time.Second)
	params := postgresOrderParams("AHS-74539", "AV-20260821", departure)
	params.Passengers = append(params.Passengers, postgresPassenger("Анна", "Смирнова", "729876543"))
	secondTicket := params.Tickets[0]
	secondTicket.TicketNumber = "AB-20269999"
	secondTicket.PassengerIndex = 1
	params.Tickets = append(params.Tickets, secondTicket)
	params.TotalPrice = 20000
	params.CurrentTotalPrice = 20000
	params.PayableAmount = 20000
	params.BonusEarned = 400

	created, err := ordersRepository.Create(ctx, params)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if len(created.Tickets) != 2 {
		t.Fatalf("expected two passenger tickets, got %d", len(created.Tickets))
	}
	if _, err := db.ExecContext(
		ctx, "UPDATE orders SET status = $1, paid_at = NOW() WHERE id = $2",
		orders.OrderStatusPaid, created.ID,
	); err != nil {
		t.Fatalf("mark order paid: %v", err)
	}
	if _, err := db.ExecContext(
		ctx, "UPDATE tickets SET status = $1 WHERE order_id = $2",
		orders.TicketStatusPaid, created.ID,
	); err != nil {
		t.Fatalf("mark tickets paid: %v", err)
	}

	var scheduledTripCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM scheduled_trips").Scan(&scheduledTripCount); err != nil {
		t.Fatalf("count scheduled trips: %v", err)
	}
	if scheduledTripCount != 2 {
		t.Fatalf("same physical trip was duplicated per passenger: %d rows", scheduledTripCount)
	}

	bookings := bookingrepo.NewRepository(db)
	loaded, err := bookings.PublicByTicket(ctx, params.Tickets[0].TicketNumber)
	if err != nil {
		t.Fatalf("public booking lookup: %v", err)
	}
	if loaded.OrderNumber != params.OrderNumber || len(loaded.Tickets) != 2 || len(loaded.Tickets[0].Segments) != 2 {
		t.Fatalf("unexpected booking graph: %+v", loaded)
	}
	if _, err := bookings.GuestByLocator(ctx, bookingaccess.Locator{OrderNumber: params.OrderNumber}, "wrong@example.com"); !errors.Is(err, bookingaccess.ErrNotFound) {
		t.Fatalf("mismatched guest email must not authorize lookup: %v", err)
	}

	publicTrips := triprepo.NewRepository(db)
	byTicket, err := publicTrips.ByTicketNumber(ctx, params.Tickets[0].TicketNumber)
	if err != nil || len(byTicket) != 2 || byTicket[0].RouteNumber != "SU 12345" ||
		byTicket[1].RouteNumber != "TK 54321" {
		t.Fatalf("trip lookup by ticket: trips=%+v err=%v", byTicket, err)
	}
	from := time.Date(departure.Year(), departure.Month(), departure.Day(), 0, 0, 0, 0, time.UTC)
	byRoute, err := publicTrips.ByRouteNumber(ctx, "SU 12345", from, from.AddDate(0, 0, 1))
	if err != nil || len(byRoute) != 1 {
		t.Fatalf("trip lookup by route: trips=%+v err=%v", byRoute, err)
	}

	stale := postgresOrderParams("OLD-10000", "CD-20260000", departure.Add(time.Hour))
	stale.GuestPaymentToken = "a38ebfda-a3f5-4b39-99e8-3eaa8089c317"
	stale.ExpiresAt = time.Now().UTC().Add(-48 * time.Hour)
	stale.Tickets[0].Segments[0].RouteNumber = "SU 00001"
	stale.Tickets[0].Segments[1].RouteNumber = "TK 00002"
	if _, err := ordersRepository.Create(ctx, stale); err != nil {
		t.Fatalf("create stale order: %v", err)
	}
	if _, err := db.ExecContext(
		ctx, "UPDATE scheduled_trips SET created_at = $1 WHERE route_number IN ($2, $3)",
		time.Now().UTC().Add(-48*time.Hour), "SU 00001", "TK 00002",
	); err != nil {
		t.Fatalf("age stale scheduled trip: %v", err)
	}
	deleted, err := ordersRepository.DeleteExpiredOrders(ctx, time.Now().UTC().Add(-24*time.Hour), 1)
	if err != nil || deleted != 1 {
		t.Fatalf("cleanup expired order: deleted=%d err=%v", deleted, err)
	}
	deletedTrips, err := ordersRepository.DeleteOrphanTrips(ctx, time.Now().UTC().Add(-24*time.Hour), 1)
	if err != nil || deletedTrips != 1 {
		t.Fatalf("cleanup first orphan trip batch: deleted=%d err=%v", deletedTrips, err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM scheduled_trips").Scan(&scheduledTripCount); err != nil {
		t.Fatalf("count scheduled trips after cleanup: %v", err)
	}
	if scheduledTripCount != 3 {
		t.Fatalf("orphan cleanup ignored batch limit: %d rows", scheduledTripCount)
	}
	deleted, err = ordersRepository.DeleteExpiredOrders(ctx, time.Now().UTC().Add(-24*time.Hour), 1)
	if err != nil || deleted != 0 {
		t.Fatalf("continue orphan cleanup without stale orders: deleted=%d err=%v", deleted, err)
	}
	deletedTrips, err = ordersRepository.DeleteOrphanTrips(ctx, time.Now().UTC().Add(-24*time.Hour), 1)
	if err != nil || deletedTrips != 1 {
		t.Fatalf("cleanup second orphan trip batch: deleted=%d err=%v", deletedTrips, err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM scheduled_trips").Scan(&scheduledTripCount); err != nil {
		t.Fatalf("count scheduled trips after second cleanup: %v", err)
	}
	if scheduledTripCount != 2 {
		t.Fatalf("cleanup removed a paid trip or kept an orphan: %d rows", scheduledTripCount)
	}
}

func postgresOrderParams(orderNumber string, ticketNumber string, departure time.Time) orders.CreateOrderParams {
	policy, _ := fare.Policy(fare.Flexible)
	return orders.CreateOrderParams{
		OrderNumber: orderNumber, GuestEmail: "guest@example.com",
		GuestPaymentToken: "b59d8297-8893-4f70-9779-dabf150627b8",
		TotalPrice:        10000, CurrentTotalPrice: 10000,
		BonusSpent: 0, BonusEarned: 200, PayableAmount: 10000,
		ExpiresAt:  departure.Add(-time.Hour),
		Passengers: []orders.OrderPassengerDraft{postgresPassenger("Иван", "Петров", "721234567")},
		Tickets: []orders.TicketDraft{{
			TicketNumber: ticketNumber, PassengerIndex: 0,
			SupplierCode: "atlas", SupplierOfferID: "81fb6f4d-d0d9-45e1-a94d-4d7843445a22",
			FareType: fare.Flexible, RefundPolicy: policy, RefundPolicyVersion: 1,
			Transport: "avia", IsInternational: true, Price: 10000,
			Segments: []orders.TicketSegmentSnapshot{{
				Order: 1, FromCity: "Москва", ToCity: "Стамбул",
				DepartureTime: departure.Format(time.RFC3339),
				ArrivalTime:   departure.Add(3 * time.Hour).Format(time.RFC3339),
				Carrier:       "Aeroflot", CarrierCode: "SU", RouteNumber: "SU 12345",
			}, {
				Order: 2, FromCity: "Стамбул", ToCity: "Париж",
				DepartureTime: departure.Add(5 * time.Hour).Format(time.RFC3339),
				ArrivalTime:   departure.Add(9 * time.Hour).Format(time.RFC3339),
				Carrier:       "Turkish Airlines", CarrierCode: "TK", RouteNumber: "TK 54321",
			}},
		}},
	}
}

func postgresPassenger(firstName string, lastName string, documentNumber string) orders.OrderPassengerDraft {
	return orders.OrderPassengerDraft{
		Source: passengers.SourceNew,
		Passenger: orders.PassengerSnapshot{
			FirstName: firstName, LastName: lastName, BirthDate: "1990-01-01", IsRussian: true,
		},
		Document: orders.DocumentSnapshot{
			Type: "international_passport", Number: documentNumber,
			VerificationStatus: "verified", LastCheckedAt: time.Now().UTC().Format(time.RFC3339),
		},
		DocumentFingerprint: strings.Repeat("a", 64),
	}
}
