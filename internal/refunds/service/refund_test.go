package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/fare"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

const (
	refundTestKey = "11111111-1111-4111-8111-111111111111"
	refundTestID  = "33333333-3333-4333-8333-333333333333"
)

func TestQuoteCalculatesRetentionFromFullTicketPrice(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	repo, gateway, service, userID := refundFixture(now)
	quote, err := service.Quote(context.Background(), &userID, repo.order.ID, refunds.Input{TicketIDs: []int{repo.tickets[0].ID}})
	if err != nil {
		t.Fatalf("Quote() error = %v", err)
	}
	if !quote.Refundable || quote.CashAmount != 300 || quote.BonusRestored != 400 || quote.BonusRevoked != 20 {
		t.Fatalf("quote = %+v", quote)
	}
	if gateway.quoteCalls != 1 || quote.Items[0].GrossRefundAmount != 700 {
		t.Fatalf("gateway calls = %d item = %+v", gateway.quoteCalls, quote.Items[0])
	}
}

func TestRefundFinancialsRestoreOnlyBonusAboveNewLimit(t *testing.T) {
	userID := 7
	financials, err := calculateRefundFinancials(&refunds.OrderData{
		ID: 5, UserID: &userID, Status: orders.OrderStatusPaid,
		CurrentTotalPrice: 10000, BonusSpent: 5000, BonusEarned: 200, PayableAmount: 5000,
	}, 8000, 7200)
	if err != nil {
		t.Fatalf("calculateRefundFinancials() error = %v", err)
	}
	if financials.currentTotalPrice != 2000 || financials.bonusSpent != 1000 ||
		financials.bonusEarned != 40 || financials.payableAmount != 1800 ||
		financials.cashAmount != 3200 || financials.bonusRestored != 4000 ||
		financials.bonusRevoked != 160 {
		t.Fatalf("financials = %+v", financials)
	}
}

func TestRefundFinancialsKeepBonusWithinNewLimit(t *testing.T) {
	userID := 7
	financials, err := calculateRefundFinancials(&refunds.OrderData{
		ID: 5, UserID: &userID, Status: orders.OrderStatusPaid,
		CurrentTotalPrice: 10000, BonusSpent: 2000, BonusEarned: 200, PayableAmount: 8000,
	}, 2000, 1800)
	if err != nil {
		t.Fatalf("calculateRefundFinancials() error = %v", err)
	}
	if financials.currentTotalPrice != 8000 || financials.bonusSpent != 2000 ||
		financials.bonusEarned != 160 || financials.payableAmount != 6200 ||
		financials.cashAmount != 1800 || financials.bonusRestored != 0 ||
		financials.bonusRevoked != 40 {
		t.Fatalf("financials = %+v", financials)
	}
}

func TestRefundFinancialsDistinguishLocalStateFromSupplierResponse(t *testing.T) {
	userID := 7
	order := &refunds.OrderData{
		ID: 5, UserID: &userID, Status: orders.OrderStatusPaid,
		CurrentTotalPrice: 10000, BonusSpent: 2000, BonusEarned: 199, PayableAmount: 8000,
	}

	_, err := calculateRefundFinancials(order, 2000, 1800)
	if !errors.Is(err, refunds.ErrInvalidFinancialState) {
		t.Fatalf("local state error = %v, want ErrInvalidFinancialState", err)
	}

	order.BonusEarned = 200
	_, err = calculateRefundFinancials(order, 2000, 2001)
	if !errors.Is(err, refunds.ErrSupplierMismatch) {
		t.Fatalf("supplier response error = %v, want ErrSupplierMismatch", err)
	}
}

func TestOrderStatusAfterRefund(t *testing.T) {
	status, err := orderStatusAfterRefund(3, 1)
	if err != nil || status != orders.OrderStatusPartiallyRefunded {
		t.Fatalf("partial status = %q, error = %v", status, err)
	}

	status, err = orderStatusAfterRefund(3, 3)
	if err != nil || status != orders.OrderStatusRefunded {
		t.Fatalf("full status = %q, error = %v", status, err)
	}

	if _, err = orderStatusAfterRefund(3, 4); !errors.Is(err, refunds.ErrInvalidFinancialState) {
		t.Fatalf("invalid counts error = %v, want ErrInvalidFinancialState", err)
	}
}

func TestMapErrorKeepsLocalAndSupplierFailuresSeparate(t *testing.T) {
	if got := mapError(refunds.ErrInvalidFinancialState); !errors.Is(got, apierror.ErrInternal) ||
		!errors.Is(apierror.Cause(got), refunds.ErrInvalidFinancialState) {
		t.Fatalf("local financial error = %v, want internal error", got)
	}
	if got := mapError(refunds.ErrSupplierMismatch); !errors.Is(got, errSupplierResponse) ||
		!errors.Is(apierror.Cause(got), refunds.ErrSupplierMismatch) {
		t.Fatalf("supplier mismatch error = %v, want supplier response error", got)
	}
}

func TestRefundRetriesProcessingOperationWithSameKey(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	repo, gateway, service, userID := refundFixture(now)
	gateway.executeErr = context.DeadlineExceeded
	input := refunds.Input{TicketIDs: []int{repo.tickets[0].ID}, IdempotencyKey: refundTestKey}

	processing, err := service.Refund(context.Background(), &userID, repo.order.ID, input)
	if err != nil {
		t.Fatalf("first Refund() error = %v", err)
	}
	if processing.Status != refunds.StatusProcessing || processing.AttemptCount != 1 ||
		repo.tickets[0].Status != orders.TicketStatusRefundPending || repo.createCalls != 1 {
		t.Fatalf("processing result = %+v ticket=%s creates=%d", processing, repo.tickets[0].Status, repo.createCalls)
	}

	gateway.executeErr = nil
	tooEarly, err := service.Refund(context.Background(), &userID, repo.order.ID, input)
	if err != nil || tooEarly.Status != refunds.StatusProcessing || gateway.executeCalls != 1 {
		t.Fatalf("early retry = %+v error=%v executeCalls=%d", tooEarly, err, gateway.executeCalls)
	}
	service.now = func() time.Time { return now.Add(16 * time.Second) }
	completed, err := service.Refund(context.Background(), &userID, repo.order.ID, input)
	if err != nil {
		t.Fatalf("second Refund() error = %v", err)
	}
	if completed.Status != refunds.StatusSuccess || completed.SupplierRefundID != refundTestID ||
		completed.OrderStatus != orders.OrderStatusRefunded || completed.BonusDebt == nil || *completed.BonusDebt != 15 {
		t.Fatalf("completed result = %+v", completed)
	}
	if repo.createCalls != 1 || gateway.quoteCalls != 1 || gateway.executeCalls != 2 {
		t.Fatalf("calls = creates:%d quotes:%d executes:%d", repo.createCalls, gateway.quoteCalls, gateway.executeCalls)
	}
}

func TestRefundResumesCommittedOperationAfterInitialLookupMiss(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	repo, gateway, service, userID := refundFixture(now)
	input := refunds.Input{TicketIDs: []int{repo.tickets[0].ID}, IdempotencyKey: refundTestKey}

	first, err := service.Refund(context.Background(), &userID, repo.order.ID, input)
	if err != nil {
		t.Fatalf("first Refund() error = %v", err)
	}
	repo.getByKeyMisses = 1

	repeated, err := service.Refund(context.Background(), &userID, repo.order.ID, input)
	if err != nil {
		t.Fatalf("repeated Refund() error = %v", err)
	}
	if repeated.Status != refunds.StatusSuccess || repeated.RefundID != first.RefundID {
		t.Fatalf("repeated refund = %+v, want operation %d with status %s", repeated, first.RefundID, refunds.StatusSuccess)
	}
	if repo.createCalls != 1 || gateway.quoteCalls != 1 || gateway.executeCalls != 1 {
		t.Fatalf("calls = creates:%d quotes:%d executes:%d, want 1 each", repo.createCalls, gateway.quoteCalls, gateway.executeCalls)
	}
}

func TestReconciliationStopsAfterAttemptLimit(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	repo, gateway, service, userID := refundFixture(now)
	gateway.executeErr = context.DeadlineExceeded
	input := refunds.Input{TicketIDs: []int{repo.tickets[0].ID}, IdempotencyKey: refundTestKey}
	if _, err := service.Refund(context.Background(), &userID, repo.order.ID, input); err != nil {
		t.Fatalf("initial Refund() error = %v", err)
	}

	clock := now
	service.now = func() time.Time { return clock }
	for repo.operation.Status == refunds.StatusProcessing {
		clock = repo.operation.NextRetryAt
		if err := service.Reconcile(context.Background(), 10); err != nil {
			t.Fatalf("Reconcile() error = %v", err)
		}
	}
	if repo.operation.Status != refunds.StatusRequiresReview ||
		repo.operation.AttemptCount != refunds.MaxReconciliationAttempts ||
		repo.operation.FailureCode != refunds.FailureReconciliationExhausted ||
		repo.tickets[0].Status != orders.TicketStatusRefundPending {
		t.Fatalf("review operation = %+v ticket=%s", repo.operation, repo.tickets[0].Status)
	}
	if gateway.executeCalls != refunds.MaxReconciliationAttempts {
		t.Fatalf("ExecuteRefund() calls = %d, want %d", gateway.executeCalls, refunds.MaxReconciliationAttempts)
	}
	if err := service.Reconcile(context.Background(), 10); err != nil {
		t.Fatalf("final Reconcile() error = %v", err)
	}
	if gateway.executeCalls != refunds.MaxReconciliationAttempts {
		t.Fatalf("ExecuteRefund() continued after requires_review: %d", gateway.executeCalls)
	}
}

func TestRefundRejectsSameKeyForDifferentSelection(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	repo, gateway, service, userID := refundFixture(now)
	gateway.executeErr = context.DeadlineExceeded
	first := refunds.Input{TicketIDs: []int{repo.tickets[0].ID}, IdempotencyKey: refundTestKey}
	if _, err := service.Refund(context.Background(), &userID, repo.order.ID, first); err != nil {
		t.Fatalf("first Refund() error = %v", err)
	}

	second := refunds.Input{All: true, IdempotencyKey: refundTestKey}
	_, err := service.Refund(context.Background(), &userID, repo.order.ID, second)
	var apiErr *apierror.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "idempotency_conflict" {
		t.Fatalf("second Refund() error = %v", err)
	}
	if gateway.executeCalls != 1 {
		t.Fatalf("ExecuteRefund() calls = %d, want 1", gateway.executeCalls)
	}
}

func TestRefundRestoresPaidStatusWhenSupplierRejects(t *testing.T) {
	now := time.Date(2026, time.August, 29, 12, 0, 0, 0, time.UTC)
	repo, gateway, service, userID := refundFixture(now)
	gateway.reject = true
	result, err := service.Refund(context.Background(), &userID, repo.order.ID, refunds.Input{
		TicketIDs: []int{repo.tickets[0].ID}, IdempotencyKey: refundTestKey,
	})
	if err != nil {
		t.Fatalf("Refund() error = %v", err)
	}
	if result.Status != refunds.StatusFailed || result.FailureCode != supplier.RefundReasonDeadlinePassed ||
		repo.tickets[0].Status != orders.TicketStatusPaid || repo.bonusApplied {
		t.Fatalf("failed result = %+v ticket=%s bonusApplied=%v", result, repo.tickets[0].Status, repo.bonusApplied)
	}
}

func refundFixture(now time.Time) (*memoryRefundRepository, *refundGateway, *Service, int) {
	userID := 7
	policy, _ := fare.Policy(fare.Flexible)
	repo := &memoryRefundRepository{
		order: refunds.OrderData{
			ID: 5, UserID: &userID, Status: orders.OrderStatusPaid,
			CurrentTotalPrice: 1000, BonusSpent: 400, BonusEarned: 20, PayableAmount: 600,
		},
		tickets: []refunds.TicketData{{
			ID: 11, TicketNumber: "AB-12300001", Status: orders.TicketStatusPaid,
			SupplierCode:    supplier.ProviderAtlas,
			SupplierOfferID: "22222222-2222-4222-8222-222222222222",
			FareType:        fare.Flexible, RefundPolicy: policy, RefundPolicyVersion: 1,
			Price: 1000, DepartureAt: now.Add(48 * time.Hour),
		}},
	}
	gateway := &refundGateway{}
	service := &Service{repo: repo, supplier: gateway, now: func() time.Time { return now }}
	return repo, gateway, service, userID
}

type refundGateway struct {
	quoteCalls   int
	executeCalls int
	executeErr   error
	reject       bool
}

func (g *refundGateway) SearchOffers(context.Context, supplier.SearchRequest) ([]supplier.TripOption, error) {
	return nil, errors.New("not implemented")
}

func (g *refundGateway) QuoteRefund(_ context.Context, request supplier.RefundQuoteRequest) (*supplier.RefundQuote, error) {
	g.quoteCalls++
	items := make([]supplier.RefundQuoteItemResult, 0, len(request.Items))
	for _, item := range request.Items {
		policy, _ := fare.Policy(item.FareType)
		items = append(items, supplier.RefundQuoteItemResult{
			TicketID: item.TicketID, Eligible: true, Reason: supplier.RefundReasonAllowed,
			RefundPercent: 70, GrossRefundAmount: item.GrossAmount * 70 / 100, Policy: policy,
		})
	}
	return &supplier.RefundQuote{Items: items}, nil
}

func (g *refundGateway) ExecuteRefund(_ context.Context, request supplier.ExecuteRefundRequest) (*supplier.ExecuteRefundResult, error) {
	g.executeCalls++
	if g.executeErr != nil {
		return nil, g.executeErr
	}
	items := make([]supplier.ExecuteRefundItemResult, 0, len(request.Items))
	status := supplier.RefundStatusSuccess
	failureCode := ""
	for _, item := range request.Items {
		policy, _ := fare.Policy(item.FareType)
		result := supplier.ExecuteRefundItemResult{
			TicketID: item.TicketID, Refunded: true, Reason: supplier.RefundReasonAllowed,
			RefundPercent: 70, RefundAmount: item.GrossAmount * 70 / 100, Policy: policy,
		}
		if g.reject {
			status = supplier.RefundStatusRejected
			failureCode = supplier.RefundReasonDeadlinePassed
			result.Refunded = false
			result.Reason = supplier.RefundReasonDeadlinePassed
			result.RefundPercent = 0
			result.RefundAmount = 0
		}
		items = append(items, result)
	}
	return &supplier.ExecuteRefundResult{
		SupplierRefundID: refundTestID, Status: status, FailureCode: failureCode, Items: items,
	}, nil
}

type memoryRefundRepository struct {
	order          refunds.OrderData
	tickets        []refunds.TicketData
	operation      *refunds.Operation
	getByKeyMisses int
	createCalls    int
	bonusApplied   bool
}

func (r *memoryRefundRepository) Load(context.Context, int, []int, bool) (*refunds.OrderData, []refunds.TicketData, error) {
	order := r.order
	return &order, append([]refunds.TicketData(nil), r.tickets...), nil
}

func (r *memoryRefundRepository) GetByKey(_ context.Context, key string) (*refunds.Operation, error) {
	if r.getByKeyMisses > 0 {
		r.getByKeyMisses--
		return nil, refunds.ErrNotFound
	}
	if r.operation == nil || r.operation.IdempotencyKey != key {
		return nil, refunds.ErrNotFound
	}
	return r.operation, nil
}

func (r *memoryRefundRepository) ListProcessing(_ context.Context, now time.Time, maxAttempts int, _ int) ([]refunds.Operation, error) {
	if r.operation != nil && r.operation.Status == refunds.StatusProcessing &&
		r.operation.AttemptCount < maxAttempts && r.operation.ReconciliationDeadline.After(now) &&
		!r.operation.NextRetryAt.After(now) {
		return []refunds.Operation{*r.operation}, nil
	}
	return nil, nil
}

func (r *memoryRefundRepository) MarkReviewDue(_ context.Context, now time.Time, maxAttempts int) (int, error) {
	if r.operation != nil && r.operation.Status == refunds.StatusProcessing &&
		(r.operation.AttemptCount >= maxAttempts || !r.operation.ReconciliationDeadline.After(now)) {
		r.operation.Status = refunds.StatusRequiresReview
		r.operation.FailureCode = refunds.FailureReconciliationExhausted
		return 1, nil
	}
	return 0, nil
}

func (r *memoryRefundRepository) RecordAttemptError(_ context.Context, params refunds.AttemptErrorParams) error {
	r.operation.AttemptCount++
	r.operation.LastError = params.Message
	r.operation.Status = params.Status
	r.operation.FailureCode = params.FailureCode
	r.operation.NextRetryAt = params.NextRetryAt
	return nil
}

func (r *memoryRefundRepository) WithinTransaction(ctx context.Context, operation func(refunds.Transaction) error) error {
	return operation(&memoryRefundTransaction{repository: r})
}

type memoryRefundTransaction struct {
	repository *memoryRefundRepository
}

func (t *memoryRefundTransaction) LockOrder(context.Context, int) (*refunds.OrderData, error) {
	return &t.repository.order, nil
}

func (t *memoryRefundTransaction) GetByKey(_ context.Context, key string) (*refunds.Operation, error) {
	return t.repository.GetByKey(context.Background(), key)
}

func (t *memoryRefundTransaction) LockTickets(context.Context, int, []int, bool) ([]refunds.TicketData, error) {
	return append([]refunds.TicketData(nil), t.repository.tickets...), nil
}

func (t *memoryRefundTransaction) LockOperation(context.Context, int) (*refunds.Operation, error) {
	return t.repository.operation, nil
}

func (t *memoryRefundTransaction) SuccessfulPaymentID(context.Context, int) (int, error) {
	return 17, nil
}

func (t *memoryRefundTransaction) CreateProcessing(_ context.Context, params refunds.CreateParams) (int, bool, error) {
	if t.repository.operation != nil {
		return 0, false, nil
	}
	t.repository.createCalls++
	operation := &refunds.Operation{
		ID: 23, Order: t.repository.order, PaymentID: params.PaymentID,
		Status: refunds.StatusProcessing, IdempotencyKey: params.IdempotencyKey,
		RequestHash: params.RequestHash, SupplierCode: params.SupplierCode,
		NextRetryAt: params.NextRetryAt, ReconciliationDeadline: params.ReconciliationDeadline,
		CashAmount: params.CashAmount, BonusRestored: params.BonusRestored,
		BonusRevoked: params.BonusRevoked,
	}
	for _, item := range params.Items {
		ticket := t.repository.tickets[0]
		operation.Items = append(operation.Items, refunds.OperationItem{
			TicketData: ticket, Reason: item.Reason, RefundPercent: item.RefundPercent,
			GrossAmount: item.GrossAmount, SupplierRefundAmount: item.SupplierRefundAmount,
		})
	}
	t.repository.operation = operation
	return operation.ID, true, nil
}

func (t *memoryRefundTransaction) MarkTicketsPending(context.Context, []int) error {
	t.repository.tickets[0].Status = orders.TicketStatusRefundPending
	t.repository.operation.Items[0].Status = orders.TicketStatusRefundPending
	return nil
}

func (t *memoryRefundTransaction) MarkTicketsRefunded(context.Context, []int) error {
	t.repository.tickets[0].Status = orders.TicketStatusRefunded
	t.repository.operation.Items[0].Status = orders.TicketStatusRefunded
	return nil
}

func (t *memoryRefundTransaction) CountRefundedTickets(context.Context, int) (int, int, error) {
	return 1, 1, nil
}

func (t *memoryRefundTransaction) RestoreTicketsPaid(context.Context, []int) error {
	t.repository.tickets[0].Status = orders.TicketStatusPaid
	t.repository.operation.Items[0].Status = orders.TicketStatusPaid
	return nil
}

func (t *memoryRefundTransaction) MarkSuccess(_ context.Context, params refunds.SuccessParams) error {
	operation := t.repository.operation
	operation.Status = refunds.StatusSuccess
	operation.SupplierRefundID = params.SupplierRefundID
	operation.AttemptCount++
	operation.CashAmount = params.CashAmount
	operation.BonusRestored = params.BonusRestored
	operation.BonusRevoked = params.BonusRevoked
	operation.BonusBalanceAfter = params.BonusBalanceAfter
	operation.BonusDebtAfter = params.BonusDebtAfter
	operation.Items[0].Reason = params.Items[0].Reason
	operation.Items[0].RefundPercent = params.Items[0].RefundPercent
	operation.Items[0].SupplierRefundAmount = params.Items[0].SupplierRefundAmount
	return nil
}

func (t *memoryRefundTransaction) MarkFailed(_ context.Context, params refunds.FailureParams) error {
	operation := t.repository.operation
	operation.Status = refunds.StatusFailed
	operation.SupplierRefundID = params.SupplierRefundID
	operation.FailureCode = params.FailureCode
	operation.AttemptCount++
	operation.CashAmount = 0
	operation.BonusRestored = 0
	operation.BonusRevoked = 0
	operation.Items[0].Reason = params.Items[0].Reason
	operation.Items[0].RefundPercent = 0
	operation.Items[0].SupplierRefundAmount = 0
	return nil
}

func (t *memoryRefundTransaction) UpdateOrderAfterRefund(_ context.Context, params refunds.OrderRefundParams) error {
	t.repository.order.Status = params.Status
	t.repository.order.CurrentTotalPrice = params.CurrentTotalPrice
	t.repository.order.BonusSpent = params.BonusSpent
	t.repository.order.BonusEarned = params.BonusEarned
	t.repository.order.PayableAmount = params.PayableAmount
	t.repository.operation.Order = t.repository.order
	return nil
}

func (t *memoryRefundTransaction) ApplyBonus(_ context.Context, params refunds.BonusParams) (*refunds.BonusResult, error) {
	t.repository.bonusApplied = true
	return &refunds.BonusResult{Balance: 0, Debt: 15, DebtCreated: 15}, nil
}
