package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
)

func (s *Service) Pay(ctx context.Context, userID *int, orderID int, guestPaymentToken string) (*checkout.PayOutput, error) {
	if orderID <= 0 {
		return nil, apierror.ErrInvalidRequest
	}
	success, err := s.provider.Process(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate mock payment result: %w", err)
	}
	outcome, err := s.processPayment(ctx, userID, orderID, guestPaymentToken, success)
	if err != nil {
		return nil, mapPaymentError(ctx, orderID, err)
	}
	if outcome.expired {
		return nil, mapPaymentError(ctx, orderID, orders.ErrExpired)
	}

	logPaymentResult(ctx, orderID, outcome.paymentID, success)
	return &checkout.PayOutput{
		OrderID:      orderID,
		PaymentID:    outcome.paymentID,
		Status:       payments.ResultStatus(success),
		Amount:       outcome.order.PayableAmount,
		BonusSpent:   outcome.order.BonusSpent,
		BonusEarned:  outcome.order.BonusEarned,
		BonusBalance: outcome.bonusBalance,
	}, nil
}

func logPaymentResult(ctx context.Context, orderID int, paymentID int, success bool) {
	if success {
		slog.InfoContext(ctx, "Mock-оплата успешно проведена",
			slog.Int("order_id", orderID),
			slog.Int("payment_id", paymentID),
		)
		return
	}
	slog.WarnContext(ctx, "Mock-оплата отклонена",
		slog.Int("order_id", orderID),
		slog.Int("payment_id", paymentID),
	)
}
