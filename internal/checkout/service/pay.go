package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) Pay(ctx context.Context, userID *int, orderID int, guestPaymentToken string) (*checkout.PayOutput, error) {
	if orderID <= 0 {
		return nil, apierror.ErrInvalidRequest
	}
	success, err := s.provider.Process(ctx)
	if err != nil {
		logger.Error("Ошибка генерации результата mock-оплаты: %v", err)
		return nil, err
	}
	outcome, err := s.processPayment(ctx, userID, orderID, guestPaymentToken, success)
	if err != nil {
		return nil, mapPaymentError(orderID, err)
	}
	if outcome.expired {
		return nil, mapPaymentError(orderID, orders.ErrExpired)
	}

	logPaymentResult(orderID, outcome.paymentID, success)
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

func logPaymentResult(orderID int, paymentID int, success bool) {
	if success {
		logger.Info("Mock-оплата успешно завершена: orderID=%d paymentID=%d", orderID, paymentID)
		return
	}
	logger.Warn("Mock-оплата отклонена: orderID=%d paymentID=%d", orderID, paymentID)
}
