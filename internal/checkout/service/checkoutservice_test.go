package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
)

func TestPayOrchestratesSuccessfulPayment(t *testing.T) {
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	userID := 7
	tx := &fakeTransaction{
		order: orders.PaymentData{
			ID: 42, UserID: &userID, Status: orders.OrderStatusCreated,
			PayableAmount: 5000, BonusSpent: 100, BonusEarned: 102, ExpiresAt: now.Add(time.Hour),
		},
		paymentID: 12,
		balance:   325,
	}
	repo := &fakeRepository{tx: tx}
	synchronizer := &fakeSynchronizer{calls: &tx.calls}
	svc := &Service{
		repo: repo, provider: successfulProvider(), passengers: synchronizer,
		now: func() time.Time { return now },
	}

	out, err := svc.Pay(context.Background(), &userID, 42, "")
	if err != nil {
		t.Fatalf("Pay() error = %v", err)
	}
	if out.Status != payments.StatusSuccess || out.PaymentID != 12 || out.Amount != 5000 {
		t.Fatalf("Pay() output = %+v", out)
	}
	if out.BonusBalance == nil || *out.BonusBalance != tx.balance {
		t.Fatalf("bonus balance = %v", out.BonusBalance)
	}
	if !repo.committed || repo.rolledBack {
		t.Fatalf("transaction state = committed:%t rolledBack:%t", repo.committed, repo.rolledBack)
	}
	wantCalls := []string{
		"lock_order", "create_payment", "mark_paid", "lock_user",
		"spend_bonus", "earn_bonus", "list_passengers", "sync_passengers", "get_balance",
	}
	if !reflect.DeepEqual(tx.calls, wantCalls) {
		t.Fatalf("transaction calls = %v, want %v", tx.calls, wantCalls)
	}
	if tx.paymentRecord.Status != payments.StatusSuccess || tx.paymentRecord.Provider != payments.ProviderMock {
		t.Fatalf("payment record = %+v", tx.paymentRecord)
	}
}

func TestPayCommitsProviderFailureWithoutChangingOrder(t *testing.T) {
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	tx := &fakeTransaction{
		order: orders.PaymentData{
			ID: 8, GuestToken: "guest-token", Status: orders.OrderStatusCreated,
			PayableAmount: 2500, ExpiresAt: now.Add(time.Hour),
		},
		paymentID: 13,
	}
	repo := &fakeRepository{tx: tx}
	svc := &Service{
		repo: repo, provider: failedProvider(), passengers: &fakeSynchronizer{},
		now: func() time.Time { return now },
	}

	out, err := svc.Pay(context.Background(), nil, 8, "guest-token")
	if err != nil {
		t.Fatalf("Pay() error = %v", err)
	}
	if out.Status != payments.StatusFailed || !repo.committed {
		t.Fatalf("Pay() output = %+v, committed = %t", out, repo.committed)
	}
	if want := []string{"lock_order", "create_payment"}; !reflect.DeepEqual(tx.calls, want) {
		t.Fatalf("transaction calls = %v, want %v", tx.calls, want)
	}
}

func TestPayCommitsExpiredStateBeforeReturningAPIError(t *testing.T) {
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	tx := &fakeTransaction{order: orders.PaymentData{
		ID: 9, GuestToken: "guest-token", Status: orders.OrderStatusCreated, ExpiresAt: now,
	}}
	repo := &fakeRepository{tx: tx}
	svc := &Service{
		repo: repo, provider: successfulProvider(), passengers: &fakeSynchronizer{},
		now: func() time.Time { return now },
	}

	_, err := svc.Pay(context.Background(), nil, 9, "guest-token")
	if !errors.Is(err, apierror.ErrOrderExpired) {
		t.Fatalf("Pay() error = %v, want expired", err)
	}
	if !repo.committed || repo.rolledBack {
		t.Fatalf("expired transaction = committed:%t rolledBack:%t", repo.committed, repo.rolledBack)
	}
	if want := []string{"lock_order", "mark_expired"}; !reflect.DeepEqual(tx.calls, want) {
		t.Fatalf("transaction calls = %v, want %v", tx.calls, want)
	}
}

func TestPayRollsBackBusinessAndPersistenceFailures(t *testing.T) {
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	userID := 7
	tests := []struct {
		name        string
		order       orders.PaymentData
		lockErr     error
		spendErr    error
		requestUser *int
		guestToken  string
		want        error
	}{
		{name: "not found", lockErr: orders.ErrNotFound, want: apierror.ErrNotFound},
		{
			name: "forbidden", requestUser: intPointer(8), want: apierror.ErrForbidden,
			order: orders.PaymentData{ID: 1, UserID: &userID, Status: orders.OrderStatusCreated, ExpiresAt: now.Add(time.Hour)},
		},
		{
			name: "cannot pay", requestUser: &userID, want: apierror.ErrInvalidRequest,
			order: orders.PaymentData{ID: 1, UserID: &userID, Status: orders.OrderStatusPaid, ExpiresAt: now.Add(time.Hour)},
		},
		{
			name: "insufficient bonus", requestUser: &userID, spendErr: bonus.ErrInsufficientBalance,
			want: apierror.ErrInsufficientBonus,
			order: orders.PaymentData{
				ID: 1, UserID: &userID, Status: orders.OrderStatusCreated,
				BonusSpent: 100, ExpiresAt: now.Add(time.Hour),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := &fakeTransaction{order: test.order, lockErr: test.lockErr, spendErr: test.spendErr, paymentID: 1}
			repo := &fakeRepository{tx: tx}
			svc := &Service{
				repo: repo, provider: successfulProvider(), passengers: &fakeSynchronizer{},
				now: func() time.Time { return now },
			}

			_, err := svc.Pay(context.Background(), test.requestUser, 1, test.guestToken)
			if !errors.Is(err, test.want) {
				t.Fatalf("Pay() error = %v, want %v", err, test.want)
			}
			if repo.committed || !repo.rolledBack {
				t.Fatalf("failed transaction = committed:%t rolledBack:%t", repo.committed, repo.rolledBack)
			}
		})
	}
}

func TestPayRejectsInvalidOrderBeforeCallingProvider(t *testing.T) {
	providerCalled := false
	svc := &Service{
		repo: &fakeRepository{tx: &fakeTransaction{}},
		provider: payments.ProviderFunc(func(context.Context) (bool, error) {
			providerCalled = true
			return true, nil
		}),
		passengers: &fakeSynchronizer{},
		now:        time.Now,
	}
	_, err := svc.Pay(context.Background(), nil, 0, "")
	if !errors.Is(err, apierror.ErrInvalidRequest) {
		t.Fatalf("Pay() error = %v, want invalid request", err)
	}
	if providerCalled {
		t.Fatal("payment provider was called for invalid order")
	}
}

func successfulProvider() payments.Provider {
	return payments.ProviderFunc(func(context.Context) (bool, error) { return true, nil })
}

func failedProvider() payments.Provider {
	return payments.ProviderFunc(func(context.Context) (bool, error) { return false, nil })
}

func intPointer(value int) *int { return &value }

type fakeRepository struct {
	tx         *fakeTransaction
	beginErr   error
	committed  bool
	rolledBack bool
}

func (f *fakeRepository) WithinTransaction(_ context.Context, operation func(checkout.Transaction) error) error {
	if f.beginErr != nil {
		return f.beginErr
	}
	if err := operation(f.tx); err != nil {
		f.rolledBack = true
		return err
	}
	f.committed = true
	return nil
}

type fakeSynchronizer struct {
	calls *[]string
	err   error
}

func (f *fakeSynchronizer) SyncAfterPayment(
	context.Context,
	passengers.PaymentTransaction,
	documents.PaymentTransaction,
	int,
	[]passengers.PaymentPassenger,
) error {
	if f.calls != nil {
		*f.calls = append(*f.calls, "sync_passengers")
	}
	return f.err
}

type fakeTransaction struct {
	order             orders.PaymentData
	lockErr           error
	paymentID         int
	paymentRecord     payments.Record
	createPaymentErr  error
	markExpiredErr    error
	markPaidErr       error
	lockUserErr       error
	spendErr          error
	earnErr           error
	listPassengersErr error
	passengers        []passengers.PaymentPassenger
	balance           int
	balanceErr        error
	calls             []string
}

func (f *fakeTransaction) call(name string) { f.calls = append(f.calls, name) }

func (f *fakeTransaction) LockForPayment(context.Context, int) (orders.PaymentData, error) {
	f.call("lock_order")
	return f.order, f.lockErr
}

func (f *fakeTransaction) Create(_ context.Context, record payments.Record) (int, error) {
	f.call("create_payment")
	f.paymentRecord = record
	return f.paymentID, f.createPaymentErr
}

func (f *fakeTransaction) MarkExpired(context.Context, int) error {
	f.call("mark_expired")
	return f.markExpiredErr
}

func (f *fakeTransaction) MarkPaid(context.Context, int) error {
	f.call("mark_paid")
	return f.markPaidErr
}

func (f *fakeTransaction) LockUserEffects(context.Context, int) error {
	f.call("lock_user")
	return f.lockUserErr
}

func (f *fakeTransaction) Spend(context.Context, int, int, int) error {
	f.call("spend_bonus")
	return f.spendErr
}

func (f *fakeTransaction) Earn(context.Context, int, int, int) error {
	f.call("earn_bonus")
	return f.earnErr
}

func (f *fakeTransaction) GetBalance(context.Context, int) (int, error) {
	f.call("get_balance")
	return f.balance, f.balanceErr
}

func (f *fakeTransaction) ListPassengersForPayment(context.Context, int) ([]passengers.PaymentPassenger, error) {
	f.call("list_passengers")
	return f.passengers, f.listPassengersErr
}

func (f *fakeTransaction) FindByDocument(context.Context, int, string) (int, error) {
	return 0, nil
}

func (f *fakeTransaction) InsertFromPayment(context.Context, int, passengers.PaymentSnapshot) (int, error) {
	return 0, nil
}

func (f *fakeTransaction) UpdateFromPayment(context.Context, int, int, passengers.PaymentSnapshot) error {
	return nil
}

func (f *fakeTransaction) UpsertUserAfterPayment(context.Context, int, documents.PaymentDocument) error {
	return nil
}

func (f *fakeTransaction) UpdatePassengerAfterPayment(context.Context, int, documents.PaymentDocument) (bool, error) {
	return false, nil
}

func (f *fakeTransaction) UpsertPassengerAfterPayment(context.Context, int, documents.PaymentDocument) error {
	return nil
}
