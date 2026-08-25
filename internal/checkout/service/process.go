package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
)

type paymentOutcome struct {
	paymentID    int
	order        orders.PaymentData
	bonusBalance *int
	expired      bool
}

func (s *Service) processPayment(
	ctx context.Context,
	userID *int,
	orderID int,
	guestPaymentToken string,
	success bool,
) (*paymentOutcome, error) {
	outcome := &paymentOutcome{}
	err := s.repo.WithinTransaction(ctx, func(tx checkout.Transaction) error {
		order, err := tx.LockForPayment(ctx, orderID)
		if err != nil {
			return err
		}
		outcome.order = order
		if err := authorizePayment(order, userID, guestPaymentToken); err != nil {
			return err
		}
		if err := ensureOrderPayable(order); err != nil {
			return err
		}
		if !s.now().UTC().Before(order.ExpiresAt) {
			outcome.expired = true
			return tx.MarkExpired(ctx, orderID)
		}

		outcome.paymentID, err = tx.Create(ctx, payments.Record{
			OrderID: orderID, Amount: order.PayableAmount,
			Status: payments.ResultStatus(success), Provider: payments.ProviderMock,
		})
		if err != nil || !success {
			return err
		}
		outcome.bonusBalance, err = s.completePayment(ctx, tx, order)
		return err
	})
	if err != nil {
		return nil, err
	}
	return outcome, nil
}
