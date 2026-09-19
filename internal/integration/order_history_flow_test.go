//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	orderrepository "github.com/Halturshik/TicketAgregator-API/internal/orders/repository"
)

func TestOrderHistoryLoadsFiveOrdersWithTickets(t *testing.T) {
	db := openIntegrationDB(t)
	repo := orderrepository.NewRepository(db)
	userID := insertUser(t, db, "history@example.com", 0)
	otherUserID := insertUser(t, db, "other-history@example.com", 0)

	for index := 0; index < 7; index++ {
		createUserOrder(
			t, repo, userID, 2, string(rune('A'+index)),
			4000, 0, 80, time.Now().UTC().Add(15*time.Minute),
		)
	}
	createUserOrder(t, repo, otherUserID, 1, "Z", 1000, 0, 20, time.Now().UTC().Add(15*time.Minute))
	markOrdersPaid(t, db)

	page, err := repo.List(context.Background(), userID, orders.ListFilter{Limit: 5})
	if err != nil {
		t.Fatalf("list first history page: %v", err)
	}
	if page.Total != 7 || page.Offset != 0 || page.Limit != 5 || len(page.Items) != 5 {
		t.Fatalf("first page = total:%d offset:%d limit:%d items:%d", page.Total, page.Offset, page.Limit, len(page.Items))
	}
	for _, item := range page.Items {
		if len(item.Tickets) != 4 || len(item.Directions) != 2 {
			t.Fatalf("order %d = tickets:%d directions:%d", item.ID, len(item.Tickets), len(item.Directions))
		}
		if item.OrderNumber == "" {
			t.Fatalf("order %d has empty order number", item.ID)
		}
		if item.CurrentTotalPrice != 4000 {
			t.Fatalf("order %d current total = %d, want 4000", item.ID, item.CurrentTotalPrice)
		}
		for _, ticket := range item.Tickets {
			if ticket.TicketNumber == "" || ticket.Passenger.FirstName == "" || ticket.Document.Number == "" {
				t.Fatalf("incomplete ticket history: %+v", ticket)
			}
			if ticket.Direction.FromCity == "" || ticket.Direction.ToCity == "" || ticket.Direction.DepartureTime == "" {
				t.Fatalf("incomplete ticket direction: %+v", ticket.Direction)
			}
		}
	}

	page, err = repo.List(context.Background(), userID, orders.ListFilter{Limit: 5, Offset: 5})
	if err != nil {
		t.Fatalf("list second history page: %v", err)
	}
	if page.Total != 7 || len(page.Items) != 2 {
		t.Fatalf("second page = total:%d items:%d", page.Total, len(page.Items))
	}
}

func TestOrderHistoryFiltersTransport(t *testing.T) {
	db := openIntegrationDB(t)
	repo := orderrepository.NewRepository(db)
	userID := insertUser(t, db, "history-filter@example.com", 0)

	avia := orderParams(&userID, "", "", 1, "A", 1000, 0, 20, time.Now().UTC().Add(15*time.Minute))
	if _, err := repo.Create(context.Background(), avia); err != nil {
		t.Fatalf("create avia order: %v", err)
	}
	rail := orderParams(&userID, "", "", 1, "R", 1200, 0, 24, time.Now().UTC().Add(15*time.Minute))
	for index := range rail.Tickets {
		rail.Tickets[index].Transport = "rail"
		rail.Tickets[index].TicketNumber = fmt.Sprintf("9R%04dA", index+1)
	}
	if _, err := repo.Create(context.Background(), rail); err != nil {
		t.Fatalf("create rail order: %v", err)
	}
	markOrdersPaid(t, db)

	page, err := repo.List(context.Background(), userID, orders.ListFilter{Transport: "rail", Limit: 5})
	if err != nil {
		t.Fatalf("list rail history: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Tickets[0].Transport != "rail" {
		t.Fatalf("rail page = %+v", page)
	}
}

func TestDeleteExpiredRemovesOnlyStaleUnpaidOrders(t *testing.T) {
	db := openIntegrationDB(t)
	repo := orderrepository.NewRepository(db)
	userID := insertUser(t, db, "cleanup-history@example.com", 0)
	now := time.Now().UTC()

	stale := createUserOrder(t, repo, userID, 1, "S", 1000, 0, 20, now.Add(-48*time.Hour))
	active := createUserOrder(t, repo, userID, 1, "A", 1000, 0, 20, now.Add(15*time.Minute))
	paid := createUserOrder(t, repo, userID, 1, "P", 1000, 0, 20, now.Add(-48*time.Hour))
	if _, err := db.Exec(`UPDATE orders SET status = 'paid', paid_at = NOW() WHERE id = $1`, paid.ID); err != nil {
		t.Fatalf("mark retained order paid: %v", err)
	}
	if _, err := db.Exec(`UPDATE tickets SET status = 'paid' WHERE order_id = $1`, paid.ID); err != nil {
		t.Fatalf("mark retained tickets paid: %v", err)
	}

	deleted, err := repo.DeleteExpiredOrders(context.Background(), now.Add(-24*time.Hour), 500)
	if err != nil {
		t.Fatalf("delete expired orders: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM orders WHERE id = $1`, stale.ID, 0)
	assertCount(t, db, `SELECT COUNT(*) FROM tickets WHERE order_id = $1`, stale.ID, 0)
	assertCount(t, db, `SELECT COUNT(*) FROM orders WHERE id = $1`, active.ID, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM orders WHERE id = $1`, paid.ID, 1)
}

func markOrdersPaid(t *testing.T, db interface {
	Exec(string, ...any) (sql.Result, error)
}) {
	t.Helper()
	if _, err := db.Exec(`UPDATE orders SET status = 'paid', paid_at = NOW()`); err != nil {
		t.Fatalf("mark history orders paid: %v", err)
	}
	if _, err := db.Exec(`UPDATE tickets SET status = 'paid'`); err != nil {
		t.Fatalf("mark history tickets paid: %v", err)
	}
}
