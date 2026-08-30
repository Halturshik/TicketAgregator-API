package refunds

import (
	"context"
	"time"
)

type Service interface {
	Quote(ctx context.Context, userID *int, orderID int, input Input) (*QuoteOutput, error)
	Refund(ctx context.Context, userID *int, orderID int, input Input) (*Result, error)
	Reconcile(ctx context.Context, limit int) error
}

type Repository interface {
	Load(ctx context.Context, orderID int, ticketIDs []int, all bool) (*OrderData, []TicketData, error)
	GetByKey(ctx context.Context, idempotencyKey string) (*Operation, error)
	ListProcessing(ctx context.Context, now time.Time, maxAttempts int, limit int) ([]Operation, error)
	MarkReviewDue(ctx context.Context, now time.Time, maxAttempts int) (int, error)
	RecordAttemptError(ctx context.Context, params AttemptErrorParams) error
	WithinTransaction(ctx context.Context, operation func(Transaction) error) error
}

type Transaction interface {
	LockOrder(ctx context.Context, orderID int) (*OrderData, error)
	GetByKey(ctx context.Context, idempotencyKey string) (*Operation, error)
	LockTickets(ctx context.Context, orderID int, ticketIDs []int, all bool) ([]TicketData, error)
	LockOperation(ctx context.Context, refundID int) (*Operation, error)
	SuccessfulPaymentID(ctx context.Context, orderID int) (int, error)
	CreateProcessing(ctx context.Context, params CreateParams) (int, bool, error)
	MarkTicketsPending(ctx context.Context, ticketIDs []int) error
	MarkTicketsRefunded(ctx context.Context, ticketIDs []int) error
	RestoreTicketsPaid(ctx context.Context, ticketIDs []int) error
	MarkSuccess(ctx context.Context, params SuccessParams) error
	MarkFailed(ctx context.Context, params FailureParams) error
	UpdateOrderStatus(ctx context.Context, orderID int) (string, error)
	ApplyBonus(ctx context.Context, params BonusParams) (*BonusResult, error)
}
