//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	checkoutrepository "github.com/Halturshik/TicketAgregator-API/internal/checkout/repository"
	checkoutservice "github.com/Halturshik/TicketAgregator-API/internal/checkout/service"
	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	orderrepository "github.com/Halturshik/TicketAgregator-API/internal/orders/repository"
	passengerrepository "github.com/Halturshik/TicketAgregator-API/internal/passengers/repository"
	passengerservice "github.com/Halturshik/TicketAgregator-API/internal/passengers/service"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	refundrepository "github.com/Halturshik/TicketAgregator-API/internal/refunds/repository"
	refundservice "github.com/Halturshik/TicketAgregator-API/internal/refunds/service"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRefundCreatesBonusDebtAndCommitsEverythingAtomically(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()
	userID := insertUser(t, db, "refund-debt@example.com", 0)
	params := orderParams(&userID, "", "", 1, "R", 5000, 0, 100, time.Now().UTC().Add(15*time.Minute))
	moveTicketDepartures(params.Tickets, 10*24*time.Hour)
	order, err := orderrepository.NewRepository(db).Create(ctx, params)
	if err != nil {
		t.Fatalf("create refundable order: %v", err)
	}
	passengerService := passengerservice.NewService(passengerrepository.NewRepository(db))
	checkoutService := checkoutservice.NewService(
		checkoutrepository.NewRepository(db), successfulPaymentProvider(), passengerService,
	)
	if _, err := checkoutService.Pay(ctx, &userID, order.ID, ""); err != nil {
		t.Fatalf("pay refundable order: %v", err)
	}
	if _, err := db.Exec(`UPDATE users SET bonus_points = 10 WHERE id = $1`, userID); err != nil {
		t.Fatalf("spend earned points fixture: %v", err)
	}

	service := refundservice.NewService(refundrepository.NewRepository(db), newSupplierGateway(t, db))
	input := refunds.Input{
		TicketIDs:      []int{order.Tickets[0].ID},
		IdempotencyKey: "11111111-1111-4111-8111-111111111111",
	}
	quote, err := service.Quote(ctx, &userID, order.ID, input)
	if err != nil {
		t.Fatalf("quote refund: %v", err)
	}
	if !quote.Refundable || quote.CashAmount != 4500 || quote.BonusRevoked != 100 {
		t.Fatalf("refund quote = %+v", quote)
	}
	result, err := service.Refund(ctx, &userID, order.ID, input)
	if err != nil {
		t.Fatalf("refund order: %v", err)
	}
	if result.OrderStatus != orders.OrderStatusRefunded || result.BonusBalance == nil || *result.BonusBalance != 0 || result.BonusDebt == nil || *result.BonusDebt != 90 {
		t.Fatalf("refund result = %+v", result)
	}
	assertCurrentTotal(t, db, order.ID, order.TotalPrice-refundedGross(result))

	assertStatus(t, db, order.ID, orders.OrderStatusRefunded)
	assertCount(t, db, `SELECT COUNT(*) FROM tickets WHERE order_id = $1 AND status = 'refunded'`, order.ID, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM refunds WHERE order_id = $1 AND status = 'success'`, order.ID, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM refund_items ri JOIN refunds r ON r.id = ri.refund_id WHERE r.order_id = $1`, order.ID, 1)
	var balance int
	var debt int
	if err := db.QueryRow(`SELECT bonus_points, bonus_debt FROM users WHERE id = $1`, userID).Scan(&balance, &debt); err != nil {
		t.Fatalf("load bonus state: %v", err)
	}
	if balance != 0 || debt != 90 {
		t.Fatalf("bonus state = balance:%d debt:%d", balance, debt)
	}

	replayed, err := service.Refund(ctx, &userID, order.ID, input)
	if err != nil || replayed.RefundID != result.RefundID || replayed.Status != refunds.StatusSuccess {
		t.Fatalf("replayed refund = %+v error = %v", replayed, err)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM refunds WHERE order_id = $1`, order.ID, 1)

	payOrderWithEarnedBonus(t, db, checkoutService, userID, "D", 50)
	assertBonusState(t, db, userID, 0, 40)
	payOrderWithEarnedBonus(t, db, checkoutService, userID, "E", 100)
	assertBonusState(t, db, userID, 60, 0)
}

func TestPartialRefundRebalancesOrderBonusWithoutExceedingHalf(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()
	userID := insertUser(t, db, "refund-rebalance@example.com", 5000)
	params := orderParams(&userID, "", "", 5, "B", 10000, 5000, 200, time.Now().UTC().Add(15*time.Minute))
	moveTicketDepartures(params.Tickets, 10*24*time.Hour)
	order, err := orderrepository.NewRepository(db).Create(ctx, params)
	if err != nil {
		t.Fatalf("create bonus order: %v", err)
	}
	passengerService := passengerservice.NewService(passengerrepository.NewRepository(db))
	checkoutService := checkoutservice.NewService(
		checkoutrepository.NewRepository(db), successfulPaymentProvider(), passengerService,
	)
	if _, err := checkoutService.Pay(ctx, &userID, order.ID, ""); err != nil {
		t.Fatalf("pay bonus order: %v", err)
	}

	service := refundservice.NewService(refundrepository.NewRepository(db), newSupplierGateway(t, db))
	selected := make([]int, 4)
	for index := range selected {
		selected[index] = order.Tickets[index].ID
	}
	result, err := service.Refund(ctx, &userID, order.ID, refunds.Input{
		TicketIDs: selected, IdempotencyKey: "12121212-1212-4212-8212-121212121212",
	})
	if err != nil {
		t.Fatalf("refund bonus tickets: %v", err)
	}
	if result.OrderStatus != orders.OrderStatusPartiallyRefunded ||
		result.CashAmount != 3200 || result.BonusRestored != 4000 || result.BonusRevoked != 160 {
		t.Fatalf("refund result = %+v", result)
	}
	assertOrderFinancials(t, db, order.ID, 2000, 1000, 40, 1800)
	assertBonusState(t, db, userID, 4040, 0)
}

func TestRefundOneTicketThenAllRemainingTickets(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()
	userID := insertUser(t, db, "partial-refund@example.com", 0)
	params := orderParams(&userID, "", "", 2, "P", 4000, 0, 80, time.Now().UTC().Add(15*time.Minute))
	moveTicketDepartures(params.Tickets, 10*24*time.Hour)
	order, err := orderrepository.NewRepository(db).Create(ctx, params)
	if err != nil {
		t.Fatalf("create multi-ticket order: %v", err)
	}
	passengerService := passengerservice.NewService(passengerrepository.NewRepository(db))
	checkoutService := checkoutservice.NewService(
		checkoutrepository.NewRepository(db), successfulPaymentProvider(), passengerService,
	)
	if _, err := checkoutService.Pay(ctx, &userID, order.ID, ""); err != nil {
		t.Fatalf("pay multi-ticket order: %v", err)
	}
	service := refundservice.NewService(refundrepository.NewRepository(db), newSupplierGateway(t, db))

	first, err := service.Refund(ctx, &userID, order.ID, refunds.Input{
		TicketIDs:      []int{order.Tickets[0].ID},
		IdempotencyKey: "22222222-2222-4222-8222-222222222222",
	})
	if err != nil {
		t.Fatalf("refund first ticket: %v", err)
	}
	if first.OrderStatus != orders.OrderStatusPartiallyRefunded {
		t.Fatalf("first order status = %q", first.OrderStatus)
	}
	currentTotal := order.TotalPrice - refundedGross(first)
	assertCurrentTotal(t, db, order.ID, currentTotal)
	remaining, err := service.Refund(ctx, &userID, order.ID, refunds.Input{
		All: true, IdempotencyKey: "33333333-3333-4333-8333-333333333333",
	})
	if err != nil {
		t.Fatalf("refund remaining tickets: %v", err)
	}
	if remaining.OrderStatus != orders.OrderStatusRefunded || len(remaining.Items) != len(order.Tickets)-1 {
		t.Fatalf("remaining refund = status:%s items:%d", remaining.OrderStatus, len(remaining.Items))
	}
	currentTotal -= refundedGross(remaining)
	assertCurrentTotal(t, db, order.ID, currentTotal)
	assertCount(t, db, `SELECT COUNT(*) FROM tickets WHERE order_id = $1 AND status = 'refunded'`, order.ID, len(order.Tickets))
	assertCount(t, db, `SELECT COUNT(*) FROM refunds WHERE order_id = $1`, order.ID, 2)
}

func TestConcurrentRefundCommitsExactlyOnce(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()
	userID := insertUser(t, db, "concurrent-refund@example.com", 0)
	params := orderParams(&userID, "", "", 1, "C", 1000, 0, 20, time.Now().UTC().Add(15*time.Minute))
	moveTicketDepartures(params.Tickets, 10*24*time.Hour)
	order, err := orderrepository.NewRepository(db).Create(ctx, params)
	if err != nil {
		t.Fatalf("create concurrent refund order: %v", err)
	}
	passengerService := passengerservice.NewService(passengerrepository.NewRepository(db))
	checkoutService := checkoutservice.NewService(
		checkoutrepository.NewRepository(db), successfulPaymentProvider(), passengerService,
	)
	if _, err := checkoutService.Pay(ctx, &userID, order.ID, ""); err != nil {
		t.Fatalf("pay concurrent refund order: %v", err)
	}
	service := refundservice.NewService(refundrepository.NewRepository(db), newSupplierGateway(t, db))
	input := refunds.Input{
		TicketIDs:      []int{order.Tickets[0].ID},
		IdempotencyKey: "44444444-4444-4444-8444-444444444444",
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	workers.Add(2)
	for range 2 {
		go func() {
			defer workers.Done()
			<-start
			_, err := service.Refund(ctx, &userID, order.ID, input)
			results <- err
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	succeeded := 0
	for err := range results {
		if err == nil {
			succeeded++
			continue
		}
		t.Fatalf("unexpected concurrent refund error: %v", err)
	}
	if succeeded != 2 {
		t.Fatalf("concurrent refunds = success:%d, want 2 idempotent responses", succeeded)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM refunds WHERE order_id = $1`, order.ID, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM refund_items ri JOIN refunds r ON r.id = ri.refund_id WHERE r.order_id = $1`, order.ID, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM supplier_simulator.refund_operations WHERE provider_code = 'atlas' AND $1 > 0`, order.ID, 1)
}

func TestRefundReconcileCompletesOperationAfterSupplierTimeout(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()
	userID := insertUser(t, db, "reconcile-refund@example.com", 0)
	params := orderParams(&userID, "", "", 1, "Q", 1000, 0, 20, time.Now().UTC().Add(15*time.Minute))
	moveTicketDepartures(params.Tickets, 10*24*time.Hour)
	order, err := orderrepository.NewRepository(db).Create(ctx, params)
	if err != nil {
		t.Fatalf("create reconcile order: %v", err)
	}
	passengerService := passengerservice.NewService(passengerrepository.NewRepository(db))
	checkoutService := checkoutservice.NewService(
		checkoutrepository.NewRepository(db), successfulPaymentProvider(), passengerService,
	)
	if _, err := checkoutService.Pay(ctx, &userID, order.ID, ""); err != nil {
		t.Fatalf("pay reconcile order: %v", err)
	}
	gateway := &timeoutOnceGateway{Gateway: newSupplierGateway(t, db)}
	service := refundservice.NewService(refundrepository.NewRepository(db), gateway)
	input := refunds.Input{
		TicketIDs:      []int{order.Tickets[0].ID},
		IdempotencyKey: "55555555-5555-4555-8555-555555555555",
	}

	processing, err := service.Refund(ctx, &userID, order.ID, input)
	if err != nil || processing.Status != refunds.StatusProcessing {
		t.Fatalf("processing refund = %+v error = %v", processing, err)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM tickets WHERE order_id = $1 AND status = 'refund_pending'`, order.ID, 1)
	if _, err := db.Exec(`UPDATE refunds SET next_retry_at = NOW() - INTERVAL '1 second' WHERE order_id = $1`, order.ID); err != nil {
		t.Fatalf("make refund due for reconciliation: %v", err)
	}
	if err := service.Reconcile(ctx, 10); err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	completed, err := service.Refund(ctx, &userID, order.ID, input)
	if err != nil || completed.Status != refunds.StatusSuccess || completed.AttemptCount != 2 {
		t.Fatalf("completed refund = %+v error = %v", completed, err)
	}
	assertCurrentTotal(t, db, order.ID, order.TotalPrice-refundedGross(completed))
	assertCount(t, db, `SELECT COUNT(*) FROM refunds WHERE order_id = $1 AND status = 'success'`, order.ID, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM supplier_simulator.refund_operations WHERE provider_code = 'atlas' AND $1 > 0`, order.ID, 1)
}

func TestRefundReconciliationStopsAtRequiresReview(t *testing.T) {
	db := openIntegrationDB(t)
	ctx := context.Background()
	userID := insertUser(t, db, "review-refund@example.com", 0)
	params := orderParams(&userID, "", "", 1, "V", 1000, 0, 20, time.Now().UTC().Add(15*time.Minute))
	moveTicketDepartures(params.Tickets, 10*24*time.Hour)
	order, err := orderrepository.NewRepository(db).Create(ctx, params)
	if err != nil {
		t.Fatalf("create review order: %v", err)
	}
	passengerService := passengerservice.NewService(passengerrepository.NewRepository(db))
	checkoutService := checkoutservice.NewService(
		checkoutrepository.NewRepository(db), successfulPaymentProvider(), passengerService,
	)
	if _, err := checkoutService.Pay(ctx, &userID, order.ID, ""); err != nil {
		t.Fatalf("pay review order: %v", err)
	}
	gateway := &alwaysTimeoutGateway{Gateway: newSupplierGateway(t, db)}
	service := refundservice.NewService(refundrepository.NewRepository(db), gateway)
	input := refunds.Input{
		TicketIDs:      []int{order.Tickets[0].ID},
		IdempotencyKey: "99999999-9999-4999-8999-999999999999",
	}

	processing, err := service.Refund(ctx, &userID, order.ID, input)
	if err != nil || processing.Status != refunds.StatusProcessing || processing.AttemptCount != 1 {
		t.Fatalf("initial refund = %+v error = %v", processing, err)
	}
	for attempt := 1; attempt < refunds.MaxReconciliationAttempts; attempt++ {
		if _, err := db.Exec(`UPDATE refunds SET next_retry_at = NOW() - INTERVAL '1 second' WHERE order_id = $1`, order.ID); err != nil {
			t.Fatalf("make reconciliation attempt %d due: %v", attempt+1, err)
		}
		if err := service.Reconcile(ctx, 10); err != nil {
			t.Fatalf("Reconcile() attempt %d error = %v", attempt+1, err)
		}
	}

	review, err := service.Refund(ctx, &userID, order.ID, input)
	if err != nil || review.Status != refunds.StatusRequiresReview || review.AttemptCount != refunds.MaxReconciliationAttempts {
		t.Fatalf("review refund = %+v error = %v", review, err)
	}
	if gateway.calls.Load() != refunds.MaxReconciliationAttempts {
		t.Fatalf("ExecuteRefund() calls = %d, want %d", gateway.calls.Load(), refunds.MaxReconciliationAttempts)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM refunds WHERE order_id = $1 AND status = 'requires_review' AND failure_code = 'reconciliation_exhausted'`, order.ID, 1)
	assertCount(t, db, `SELECT COUNT(*) FROM tickets WHERE order_id = $1 AND status = 'refund_pending'`, order.ID, 1)

	if err := service.Reconcile(ctx, 10); err != nil {
		t.Fatalf("Reconcile() after review error = %v", err)
	}
	if gateway.calls.Load() != refunds.MaxReconciliationAttempts {
		t.Fatalf("ExecuteRefund() calls after review = %d, want %d", gateway.calls.Load(), refunds.MaxReconciliationAttempts)
	}
}

func refundedGross(result *refunds.Result) int {
	total := 0
	for _, item := range result.Items {
		total += item.GrossAmount
	}
	return total
}

func assertOrderFinancials(
	t *testing.T,
	db *sql.DB,
	orderID int,
	currentTotal int,
	bonusSpent int,
	bonusEarned int,
	payableAmount int,
) {
	t.Helper()
	var actualCurrentTotal int
	var actualBonusSpent int
	var actualBonusEarned int
	var actualPayableAmount int
	if err := db.QueryRow(`
		SELECT current_total_price, bonus_spent, bonus_earned, payable_amount
		FROM orders
		WHERE id = $1
	`, orderID).Scan(
		&actualCurrentTotal, &actualBonusSpent, &actualBonusEarned, &actualPayableAmount,
	); err != nil {
		t.Fatalf("load order financials: %v", err)
	}
	if actualCurrentTotal != currentTotal || actualBonusSpent != bonusSpent ||
		actualBonusEarned != bonusEarned || actualPayableAmount != payableAmount {
		t.Fatalf(
			"order financials = current:%d spent:%d earned:%d payable:%d, want current:%d spent:%d earned:%d payable:%d",
			actualCurrentTotal, actualBonusSpent, actualBonusEarned, actualPayableAmount,
			currentTotal, bonusSpent, bonusEarned, payableAmount,
		)
	}
}

func TestSupplierExecuteRefundPersistsEndToEndIdempotency(t *testing.T) {
	db := openIntegrationDB(t)
	gateway := newSupplierGateway(t, db)
	now := time.Now().UTC()
	request := supplier.ExecuteRefundRequest{
		ProviderCode:   supplier.ProviderAtlas,
		IdempotencyKey: "66666666-6666-4666-8666-666666666666",
		Items: []supplier.ExecuteRefundItem{{
			TicketID: 77, TicketNumber: "AV-20260001",
			SupplierOfferID: "77777777-7777-4777-8777-777777777777",
			FareType:        fare.Flexible, DepartureUnix: now.Add(10 * 24 * time.Hour).Unix(),
			GrossAmount: 1000,
		}},
	}
	first, err := gateway.ExecuteRefund(context.Background(), request)
	if err != nil {
		t.Fatalf("first ExecuteRefund() error = %v", err)
	}
	replayed, err := gateway.ExecuteRefund(context.Background(), request)
	if err != nil || replayed.SupplierRefundID != first.SupplierRefundID {
		t.Fatalf("replayed ExecuteRefund() = %+v error = %v", replayed, err)
	}

	changed := request
	changed.Items = append([]supplier.ExecuteRefundItem(nil), request.Items...)
	changed.Items[0].GrossAmount++
	if _, err := gateway.ExecuteRefund(context.Background(), changed); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("changed command code = %s error = %v", status.Code(err), err)
	}
	request.IdempotencyKey = "88888888-8888-4888-8888-888888888888"
	if _, err := gateway.ExecuteRefund(context.Background(), request); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("second key code = %s error = %v", status.Code(err), err)
	}
}

type timeoutOnceGateway struct {
	supplier.Gateway
	failed atomic.Bool
}

func (g *timeoutOnceGateway) ExecuteRefund(ctx context.Context, request supplier.ExecuteRefundRequest) (*supplier.ExecuteRefundResult, error) {
	if g.failed.CompareAndSwap(false, true) {
		return nil, context.DeadlineExceeded
	}
	return g.Gateway.ExecuteRefund(ctx, request)
}

type alwaysTimeoutGateway struct {
	supplier.Gateway
	calls atomic.Int32
}

func (g *alwaysTimeoutGateway) ExecuteRefund(context.Context, supplier.ExecuteRefundRequest) (*supplier.ExecuteRefundResult, error) {
	g.calls.Add(1)
	return nil, context.DeadlineExceeded
}

func moveTicketDepartures(tickets []orders.TicketDraft, after time.Duration) {
	base := time.Now().UTC().Add(after).Truncate(time.Second)
	routeDepartures := make(map[string]time.Time)
	directionIndex := 0
	for ticketIndex := range tickets {
		for segmentIndex := range tickets[ticketIndex].Segments {
			segment := &tickets[ticketIndex].Segments[segmentIndex]
			departure, exists := routeDepartures[segment.RouteNumber]
			if !exists {
				departure = base.Add(time.Duration(directionIndex*24+segmentIndex*3) * time.Hour)
				routeDepartures[segment.RouteNumber] = departure
				directionIndex++
			}
			segment.DepartureTime = departure.Format(time.RFC3339)
			segment.ArrivalTime = departure.Add(2 * time.Hour).Format(time.RFC3339)
		}
	}
}

func payOrderWithEarnedBonus(
	t *testing.T,
	db *sql.DB,
	service checkout.Service,
	userID int,
	suffix string,
	earned int,
) {
	t.Helper()
	params := orderParams(&userID, "", "", 1, suffix, 1000, 0, earned, time.Now().UTC().Add(15*time.Minute))
	order, err := orderrepository.NewRepository(db).Create(context.Background(), params)
	if err != nil {
		t.Fatalf("create debt repayment order: %v", err)
	}
	if _, err := service.Pay(context.Background(), &userID, order.ID, ""); err != nil {
		t.Fatalf("pay debt repayment order: %v", err)
	}
}

func assertBonusState(t *testing.T, db *sql.DB, userID int, wantBalance int, wantDebt int) {
	t.Helper()
	var balance int
	var debt int
	if err := db.QueryRow(`SELECT bonus_points, bonus_debt FROM users WHERE id = $1`, userID).Scan(&balance, &debt); err != nil {
		t.Fatalf("load bonus state: %v", err)
	}
	if balance != wantBalance || debt != wantDebt {
		t.Fatalf("bonus state = balance:%d debt:%d, want balance:%d debt:%d", balance, debt, wantBalance, wantDebt)
	}
}
