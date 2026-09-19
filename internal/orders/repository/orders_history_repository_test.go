package repository

import (
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func TestBuildHistoryDirectionsDeduplicatesPassengerTickets(t *testing.T) {
	outbound := orders.HistoryDirection{
		FromCity: "Москва", ToCity: "Париж",
		DepartureTime: "2026-10-10T10:00:00Z", ArrivalTime: "2026-10-10T14:00:00Z",
	}
	back := orders.HistoryDirection{
		FromCity: "Париж", ToCity: "Москва",
		DepartureTime: "2026-10-20T10:00:00Z", ArrivalTime: "2026-10-20T14:00:00Z",
	}
	items := []orders.HistoryOrder{{
		Tickets: []orders.HistoryTicket{
			{ID: 1, Direction: outbound}, {ID: 2, Direction: outbound},
			{ID: 3, Direction: back}, {ID: 4, Direction: back},
		},
	}}

	buildHistoryDirections(items)

	if len(items[0].Directions) != 2 {
		t.Fatalf("direction count = %d, want 2", len(items[0].Directions))
	}
	if items[0].Directions[0] != outbound || items[0].Directions[1] != back {
		t.Fatalf("directions = %+v, want outbound and return", items[0].Directions)
	}
}

func TestUpdateHistoryDirectionUsesFirstAndLastSegment(t *testing.T) {
	direction := orders.HistoryDirection{}
	updateHistoryDirection(&direction, historyTicketRow{
		segmentOrder: 1, fromCity: "Москва", toCity: "Стамбул",
		departureTime: time.Date(2026, 10, 10, 10, 0, 0, 0, time.UTC),
		arrivalTime:   time.Date(2026, 10, 10, 13, 0, 0, 0, time.UTC),
	})
	updateHistoryDirection(&direction, historyTicketRow{
		segmentOrder: 2, fromCity: "Стамбул", toCity: "Париж",
		departureTime: time.Date(2026, 10, 10, 15, 0, 0, 0, time.UTC),
		arrivalTime:   time.Date(2026, 10, 10, 18, 0, 0, 0, time.UTC),
	})

	if direction.FromCity != "Москва" || direction.ToCity != "Париж" {
		t.Fatalf("direction = %s -> %s, want Москва -> Париж", direction.FromCity, direction.ToCity)
	}
	if direction.DepartureTime != "2026-10-10T10:00:00Z" || direction.ArrivalTime != "2026-10-10T18:00:00Z" {
		t.Fatalf("direction times = %s -> %s", direction.DepartureTime, direction.ArrivalTime)
	}
}

func TestHistoryTicketUsesPassengerAndDocumentSnapshots(t *testing.T) {
	ticket, err := historyTicket(historyTicketRow{
		ticketID: 10, ticketNumber: "AV-12345678", status: orders.TicketStatusPaid,
		transport:     "avia",
		passengerJSON: []byte(`{"first_name":"Иван","middle_name":"Иванович","last_name":"Иванов","birth_date":"1990-01-01","is_russian":true}`),
		documentJSON:  []byte(`{"type":"international_passport","number":"701234567","verification_status":"verified","last_checked_at":"2026-09-01T10:00:00Z"}`),
	})
	if err != nil {
		t.Fatalf("historyTicket() error = %v", err)
	}
	if ticket.Passenger.FirstName != "Иван" || ticket.Passenger.LastName != "Иванов" {
		t.Fatalf("passenger = %+v", ticket.Passenger)
	}
	if ticket.Document.Type != "international_passport" || ticket.Document.Number != "701234567" {
		t.Fatalf("document = %+v", ticket.Document)
	}
}
