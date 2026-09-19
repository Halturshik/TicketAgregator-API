//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	bonusrepository "github.com/Halturshik/TicketAgregator-API/internal/bonus/repository"
	checkoutrepository "github.com/Halturshik/TicketAgregator-API/internal/checkout/repository"
	checkoutservice "github.com/Halturshik/TicketAgregator-API/internal/checkout/service"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	documentrepository "github.com/Halturshik/TicketAgregator-API/internal/documents/repository"
	documentservice "github.com/Halturshik/TicketAgregator-API/internal/documents/service"
	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	orderrepository "github.com/Halturshik/TicketAgregator-API/internal/orders/repository"
	orderservice "github.com/Halturshik/TicketAgregator-API/internal/orders/service"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	passengerrepository "github.com/Halturshik/TicketAgregator-API/internal/passengers/repository"
	passengerservice "github.com/Halturshik/TicketAgregator-API/internal/passengers/service"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	searchrepository "github.com/Halturshik/TicketAgregator-API/internal/search/repository"
	searchservice "github.com/Halturshik/TicketAgregator-API/internal/search/service"
	searchstore "github.com/Halturshik/TicketAgregator-API/internal/search/store"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	userrepository "github.com/Halturshik/TicketAgregator-API/internal/users/repository"
	userservice "github.com/Halturshik/TicketAgregator-API/internal/users/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	redisclient "github.com/redis/go-redis/v9"
)

func TestTicketPurchaseFlowOnPostgres(t *testing.T) {
	db := openIntegrationDB(t)

	ctx := context.Background()
	userID := insertUser(t, db, "buyer@example.com", 500)
	otherUserID := insertUser(t, db, "other@example.com", 0)
	orderRepo := orderrepository.NewRepository(db)
	passengerSvc := passengerservice.NewService(passengerrepository.NewRepository(db))
	paymentSvc := checkoutservice.NewService(
		checkoutrepository.NewRepository(db), successfulPaymentProvider(), passengerSvc,
	)

	order := createUserOrder(t, orderRepo, userID, 2, "A", 6000, 100, 120, time.Now().UTC().Add(15*time.Minute))
	if len(order.Tickets) != 4 {
		t.Fatalf("created ticket count = %d, want 4", len(order.Tickets))
	}
	payment, err := paymentSvc.Pay(ctx, &userID, order.ID, "")
	if err != nil {
		t.Fatalf("pay order: %v", err)
	}
	if payment.PaymentID <= 0 || payment.BonusBalance == nil || *payment.BonusBalance != 520 {
		t.Fatalf("payment result = id:%d balance:%v", payment.PaymentID, payment.BonusBalance)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM tickets WHERE order_id = $1 AND status = 'paid'`, order.ID, 4)
	assertCount(t, db, `SELECT COUNT(*) FROM saved_passengers WHERE owner_user_id = $1 AND deleted_at IS NULL`, userID, 2)
	assertCount(t, db, `
		SELECT COUNT(*) FROM documents d
		JOIN saved_passengers p ON p.id = d.passenger_id
		WHERE p.owner_user_id = $1 AND d.verification_status = 'verified'
	`, userID, 2)
	assertCount(t, db, `SELECT COUNT(*) FROM bonus_transactions WHERE order_id = $1`, order.ID, 2)

	if _, err := paymentSvc.Pay(ctx, &userID, order.ID, ""); !errors.Is(err, apierror.ErrInvalidRequest) {
		t.Fatalf("second payment error = %v, want cannot be paid", err)
	}

	secondOrder := createUserOrder(t, orderRepo, userID, 1, "A", 1000, 0, 0, time.Now().UTC().Add(15*time.Minute))
	if _, err := paymentSvc.Pay(ctx, &userID, secondOrder.ID, ""); err != nil {
		t.Fatalf("pay repeated passenger order: %v", err)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM saved_passengers WHERE owner_user_id = $1 AND deleted_at IS NULL`, userID, 2)

	protectedOrder := createUserOrder(t, orderRepo, userID, 1, "B", 1000, 0, 0, time.Now().UTC().Add(15*time.Minute))
	if _, err := paymentSvc.Pay(ctx, &otherUserID, protectedOrder.ID, ""); !errors.Is(err, apierror.ErrForbidden) {
		t.Fatalf("foreign payment error = %v, want forbidden", err)
	}
	assertStatus(t, db, protectedOrder.ID, "created")

	guestToken := "2ac4b3b7-31a0-43a4-9322-c2f17232f76c"
	guestOrder, err := orderRepo.Create(ctx, orderParams(nil, "guest@example.com", guestToken, 1, "C", 1000, 0, 0, time.Now().UTC().Add(-time.Minute)))
	if err != nil {
		t.Fatalf("create expired guest order: %v", err)
	}
	if _, err := paymentSvc.Pay(ctx, nil, guestOrder.ID, guestToken); !errors.Is(err, apierror.ErrOrderExpired) {
		t.Fatalf("expired payment error = %v, want expired", err)
	}
	assertStatus(t, db, guestOrder.ID, "expired")
	assertCount(t, db, `SELECT COUNT(*) FROM tickets WHERE order_id = $1 AND status = 'expired'`, guestOrder.ID, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM payments WHERE order_id = $1`, guestOrder.ID, 0)
}

func TestSearchToPaymentServiceChainOnPostgres(t *testing.T) {
	db := openIntegrationDB(t)
	mini := miniredis.RunT(t)
	redis := redisclient.NewClient(&redisclient.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = redis.Close() })

	ctx := context.Background()
	userID := insertUser(t, db, "full-chain@example.com", 200)
	searchRepo := searchrepository.NewRepository(db)
	carriers, err := searchRepo.ListCarriers(ctx)
	if err != nil {
		t.Fatalf("list carriers: %v", err)
	}
	searchSvc, err := searchservice.NewService(
		searchRepo, searchstore.New(redis), newSupplierGateway(t, db), supplier.DefaultProviders, carriers,
	)
	if err != nil {
		t.Fatalf("create search service: %v", err)
	}
	documentSvc := documentservice.NewService(documentrepository.NewRepository(db), "integration-secret")
	passengerSvc := passengerservice.NewService(passengerrepository.NewRepository(db))
	orderSvc := orderservice.NewService(
		orderrepository.NewRepository(db), searchSvc, documentSvc, passengerSvc, bonusrepository.NewRepository(db),
	)
	paymentSvc := checkoutservice.NewService(
		checkoutrepository.NewRepository(db), successfulPaymentProvider(), passengerSvc,
	)

	outboundDate := time.Now().UTC().AddDate(0, 1, 0)
	returnDate := outboundDate.AddDate(0, 0, 10)
	_, err = searchSvc.Search(ctx, "avia", search.SearchInput{
		FromCityID: 999999, ToCityID: 16, Date: outboundDate.Format("2006-01-02"), Passengers: 1,
	}, &userID)
	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("unknown city search error = %v, want not found", err)
	}
	searchResult, err := searchSvc.Search(ctx, "avia", search.SearchInput{
		FromCityID: 1, ToCityID: 16,
		Date: outboundDate.Format("2006-01-02"), ReturnDate: returnDate.Format("2006-01-02"),
		Passengers: 2, Limit: 10,
	}, &userID)
	if err != nil {
		t.Fatalf("search tickets: %v", err)
	}
	if searchResult.Total < 52 || len(searchResult.Items) == 0 {
		t.Fatalf("search result = total:%d items:%d", searchResult.Total, len(searchResult.Items))
	}
	option := searchResult.Items[0]
	if option.Return == nil {
		t.Fatal("round-trip search returned one-way option")
	}

	order, err := orderSvc.Create(ctx, &userID, orders.CreateOrderInput{
		SearchID: searchResult.SearchID, TripOptionID: option.ID, UseBonus: 100,
		Passengers: []orders.PassengerBooking{
			foreignPassenger("JOHN", "DOE", "FP123450"),
			foreignPassenger("JANE", "DOE", "FP123451"),
		},
	})
	if err != nil {
		t.Fatalf("create order from search: %v", err)
	}
	if order.TotalPrice != option.Price || len(order.Tickets) != 4 || order.BonusSpent != 100 {
		t.Fatalf("created order = total:%d tickets:%d bonus:%d, search total:%d", order.TotalPrice, len(order.Tickets), order.BonusSpent, option.Price)
	}

	payment, err := paymentSvc.Pay(ctx, &userID, order.ID, "")
	if err != nil {
		t.Fatalf("pay full-chain order: %v", err)
	}
	wantBalance := 200 - order.BonusSpent + order.BonusEarned
	if payment.BonusBalance == nil || *payment.BonusBalance != wantBalance || payment.Amount != order.PayableAmount {
		t.Fatalf("payment = balance:%v payable:%d, want balance:%d payable:%d", payment.BonusBalance, payment.Amount, wantBalance, order.PayableAmount)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM tickets WHERE order_id = $1 AND status = 'paid'`, order.ID, 4)
	assertCount(t, db, `SELECT COUNT(*) FROM saved_passengers WHERE owner_user_id = $1 AND deleted_at IS NULL`, userID, 2)
	assertCount(t, db, `
		SELECT COUNT(*) FROM documents d
		JOIN saved_passengers p ON p.id = d.passenger_id
		WHERE p.owner_user_id = $1 AND d.verification_status = 'verified'
	`, userID, 2)

	refs := savedPassengerReferences(t, db, userID)
	savedBookings := make([]orders.PassengerBooking, 0, len(refs))
	for _, ref := range refs {
		passengerID := ref.passengerID
		documentID := ref.documentID
		savedBookings = append(savedBookings, orders.PassengerBooking{
			Source: passengers.SourceSaved, SavedPassengerID: &passengerID,
			Document: orders.DocumentBooking{SavedDocumentID: &documentID},
		})
	}
	otherUserID := insertUser(t, db, "full-chain-other@example.com", 0)
	if _, err := orderSvc.Create(ctx, &otherUserID, orders.CreateOrderInput{
		SearchID: searchResult.SearchID, TripOptionID: option.ID, Passengers: savedBookings,
	}); !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("foreign saved passenger order error = %v, want not found", err)
	}

	savedOrder, err := orderSvc.Create(ctx, &userID, orders.CreateOrderInput{
		SearchID: searchResult.SearchID, TripOptionID: option.ID, Passengers: savedBookings,
	})
	if err != nil {
		t.Fatalf("create order with saved passengers: %v", err)
	}
	if _, err := paymentSvc.Pay(ctx, &userID, savedOrder.ID, ""); err != nil {
		t.Fatalf("pay saved passenger order: %v", err)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM saved_passengers WHERE owner_user_id = $1 AND deleted_at IS NULL`, userID, 2)
}

func TestDomainErrorsAreMappedAcrossPostgresRepositories(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()
	userID := insertUser(t, db, "domain-errors@example.com", 0)

	userSvc := userservice.NewService(userrepository.NewRepository(db))
	if _, err := userSvc.GetProfile(ctx, userID+1000); !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("missing user error = %v, want not found", err)
	}

	passengerSvc := passengerservice.NewService(passengerrepository.NewRepository(db))
	if _, err := passengerSvc.GetOwned(ctx, userID, 999999); !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("missing passenger error = %v, want not found", err)
	}

	documentSvc := documentservice.NewService(documentrepository.NewRepository(db), "domain-error-secret")
	input := documents.SaveDocumentInput{Type: documents.TypeInternalPassport, Number: "1234567890"}
	if _, err := documentSvc.Create(ctx, userID, input); err != nil {
		t.Fatalf("create document: %v", err)
	}
	if _, err := documentSvc.Create(ctx, userID, input); !errors.Is(err, apierror.ErrDocumentAlreadyExists) {
		t.Fatalf("duplicate document error = %v, want already exists", err)
	}
}

func TestConcurrentPaymentCreatesOnePaymentAndPaysOrderOnce(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()
	userID := insertUser(t, db, "concurrent-payment@example.com", 0)
	order := createUserOrder(
		t, orderrepository.NewRepository(db), userID, 1, "Q", 2500, 0, 50,
		time.Now().UTC().Add(15*time.Minute),
	)
	passengerSvc := passengerservice.NewService(passengerrepository.NewRepository(db))
	checkoutSvc := checkoutservice.NewService(
		checkoutrepository.NewRepository(db), successfulPaymentProvider(), passengerSvc,
	)

	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	workers.Add(2)
	for range 2 {
		go func() {
			defer workers.Done()
			<-start
			_, err := checkoutSvc.Pay(ctx, &userID, order.ID, "")
			results <- err
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	succeeded := 0
	rejected := 0
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, apierror.ErrInvalidRequest):
			rejected++
		default:
			t.Fatalf("unexpected concurrent payment error: %v", err)
		}
	}
	if succeeded != 1 || rejected != 1 {
		t.Fatalf("concurrent results = success:%d rejected:%d, want 1 and 1", succeeded, rejected)
	}
	assertStatus(t, db, order.ID, orders.OrderStatusPaid)
	assertCount(t, db, `SELECT COUNT(*) FROM payments WHERE order_id = $1`, order.ID, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM bonus_transactions WHERE order_id = $1`, order.ID, 1)
}

func successfulPaymentProvider() payments.Provider {
	return payments.ProviderFunc(func(context.Context) (bool, error) { return true, nil })
}

type savedPassengerReference struct {
	passengerID int
	documentID  int
}

func savedPassengerReferences(t *testing.T, db *sql.DB, ownerUserID int) []savedPassengerReference {
	t.Helper()
	rows, err := db.Query(`
		SELECT p.id, d.id
		FROM saved_passengers p
		JOIN documents d ON d.passenger_id = p.id
		WHERE p.owner_user_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.id
	`, ownerUserID)
	if err != nil {
		t.Fatalf("query saved passenger references: %v", err)
	}
	defer rows.Close()
	result := make([]savedPassengerReference, 0)
	for rows.Next() {
		var ref savedPassengerReference
		if err := rows.Scan(&ref.passengerID, &ref.documentID); err != nil {
			t.Fatalf("scan saved passenger reference: %v", err)
		}
		result = append(result, ref)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate saved passenger references: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("saved passenger references = %d, want 2", len(result))
	}
	return result
}

func foreignPassenger(firstName string, lastName string, documentNumber string) orders.PassengerBooking {
	return orders.PassengerBooking{
		Source: passengers.SourceNew,
		Passenger: orders.PassengerSnapshot{
			FirstName: firstName, LastName: lastName, BirthDate: "1990-01-01", IsRussian: false,
		},
		Document: orders.DocumentBooking{Type: "foreign_passport", Number: documentNumber},
	}
}

func openIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv(testPostgresDSNEnv)
	if dsn == "" {
		t.Fatalf("%s is not configured", testPostgresDSNEnv)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
	applyMigrations(t, db)
	resetAppData(t, db)
	return db
}

func resetAppData(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		TRUNCATE TABLE
			supplier_simulator.refund_items, supplier_simulator.refund_operations,
			refund_items, refunds, bonus_transactions, payments, ticket_segments, scheduled_trips,
			tickets, order_passengers,
			orders, documents, saved_passengers, users
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("reset integration data: %v", err)
	}
}

func applyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine integration test path")
	}
	dir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	goose.SetTableName("goose_db_version")
	if err := goose.Up(db, dir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	supplierDir := filepath.Join(filepath.Dir(filename), "..", "..", "supplier-migrations")
	goose.SetTableName("supplier_goose_db_version")
	if err := goose.Up(db, supplierDir); err != nil {
		t.Fatalf("apply supplier migrations: %v", err)
	}
	goose.SetTableName("goose_db_version")
}

func insertUser(t *testing.T, db *sql.DB, email string, bonus int) int {
	t.Helper()
	var id int
	err := db.QueryRow(`
		INSERT INTO users
			(email, password_hash, first_name, last_name, birth_date, is_russian, bonus_points)
		VALUES ($1, 'hash', 'Test', 'User', '1990-01-01', TRUE, $2)
		RETURNING id
	`, email, bonus).Scan(&id)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func createUserOrder(t *testing.T, repo *orderrepository.Repository, userID int, passengers int, suffix string, total int, spent int, earned int, expiresAt time.Time) *orders.Order {
	t.Helper()
	order, err := repo.Create(context.Background(), orderParams(&userID, "", "", passengers, suffix, total, spent, earned, expiresAt))
	if err != nil {
		t.Fatalf("create user order: %v", err)
	}
	return order
}

func orderParams(userID *int, guestEmail string, guestToken string, passengerCount int, suffix string, total int, spent int, earned int, expiresAt time.Time) orders.CreateOrderParams {
	policy, _ := fare.Policy(fare.Flexible)
	passengerDrafts := make([]orders.OrderPassengerDraft, passengerCount)
	for i := range passengerDrafts {
		passengerDrafts[i] = orders.OrderPassengerDraft{
			Source: passengers.SourceNew,
			Passenger: orders.PassengerSnapshot{
				FirstName: "Passenger" + suffix + string(rune('A'+i)), LastName: "Test",
				BirthDate: "1990-01-01", IsRussian: true,
			},
			Document: orders.DocumentSnapshot{
				Type: "internal_passport", Number: "123456789" + string(rune('0'+i)),
				VerificationStatus: "verified", LastCheckedAt: "2026-08-22T12:00:00Z",
			},
			DocumentFingerprint: strings.Repeat(suffix, 63) + string(rune('A'+i)),
		}
	}
	tickets := make([]orders.TicketDraft, 0, passengerCount*2)
	directions := []struct {
		from, to int
		route    string
		date     string
	}{
		{from: 1, to: 2, route: "SU 1000" + suffix, date: "2026-09-10"},
	}
	if passengerCount == 2 {
		directions = append(directions, struct {
			from, to int
			route    string
			date     string
		}{from: 2, to: 1, route: "SU 2000" + suffix, date: "2026-09-20"})
	}
	price := total / (passengerCount * len(directions))
	for _, direction := range directions {
		for passengerIndex := range passengerDrafts {
			tickets = append(tickets, orders.TicketDraft{
				TicketNumber:   fmt.Sprintf("IT-%08d", integrationTicketSequence.Add(1)),
				PassengerIndex: passengerIndex, SupplierCode: "atlas",
				SupplierOfferID: uuid.NewSHA1(uuid.NameSpaceOID, []byte("integration-offer-"+suffix)).String(), FareType: fare.Flexible,
				RefundPolicy: policy, RefundPolicyVersion: 1,
				Transport: "avia", Price: price,
				Segments: []orders.TicketSegmentSnapshot{{
					Order: 1, FromCityID: direction.from, ToCityID: direction.to,
					FromCity: "From", ToCity: "To", DepartureTime: direction.date + "T10:00:00Z",
					ArrivalTime: direction.date + "T12:00:00Z", CarrierID: 1,
					Carrier: "Aeroflot", CarrierCode: "SU", RouteNumber: direction.route,
				}},
			})
		}
	}
	return orders.CreateOrderParams{
		OrderNumber: fmt.Sprintf("TST-%05d", integrationOrderSequence.Add(1)),
		UserID:      userID, GuestEmail: guestEmail, GuestPaymentToken: guestToken,
		TotalPrice: total, CurrentTotalPrice: total,
		BonusSpent: spent, BonusEarned: earned, PayableAmount: total - spent,
		ExpiresAt: expiresAt, Passengers: passengerDrafts, Tickets: tickets,
	}
}

var integrationTicketSequence atomic.Int64
var integrationOrderSequence atomic.Int64

func assertCount(t *testing.T, db *sql.DB, query string, id int, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(query, id).Scan(&got); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}

func assertStatus(t *testing.T, db *sql.DB, orderID int, want string) {
	t.Helper()
	var got string
	if err := db.QueryRow(`SELECT status FROM orders WHERE id = $1`, orderID).Scan(&got); err != nil {
		t.Fatalf("query order status: %v", err)
	}
	if got != want {
		t.Fatalf("order status = %q, want %q", got, want)
	}
}

func assertCurrentTotal(t *testing.T, db *sql.DB, orderID int, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(`SELECT current_total_price FROM orders WHERE id = $1`, orderID).Scan(&got); err != nil {
		t.Fatalf("query current order total: %v", err)
	}
	if got != want {
		t.Fatalf("current order total = %d, want %d", got, want)
	}
}
