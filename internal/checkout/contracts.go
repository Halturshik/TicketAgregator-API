package checkout

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
)

type Service interface {
	Pay(ctx context.Context, userID *int, orderID int, guestPaymentToken string) (*PayOutput, error)
}

type Transaction interface {
	orders.PaymentTransaction
	payments.Transaction
	bonus.PaymentTransaction
	passengers.PaymentTransaction
	documents.PaymentTransaction

	LockUserEffects(ctx context.Context, userID int) error
}

type Repository interface {
	WithinTransaction(ctx context.Context, operation func(Transaction) error) error
}
