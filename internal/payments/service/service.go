package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
	"github.com/Halturshik/TicketAgregator-API/internal/payments/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) payments.Service {
	return &Service{repo: repo}
}

func (s *Service) Pay(ctx context.Context, userID *int, orderID int, guestPaymentToken string) (*payments.PayOutput, error) {
	if orderID <= 0 {
		return nil, apierror.ErrInvalidRequest
	}
	success, err := mockSuccess()
	if err != nil {
		logger.Error("Ошибка генерации результата mock-оплаты: %v", err)
		return nil, err
	}
	paymentID, data, balance, err := s.repo.Pay(ctx, orderID, userID, guestPaymentToken, success)
	if err == sql.ErrNoRows {
		return nil, apierror.ErrNotFound
	}
	if errors.Is(err, repository.ErrPaymentForbidden) {
		return nil, apierror.ErrForbidden
	}
	if errors.Is(err, repository.ErrInsufficientBonus) {
		return nil, apierror.New("insufficient_bonus", "Недостаточно бонусов для оплаты заказа", 409)
	}
	if errors.Is(err, repository.ErrOrderCannotBePaid) {
		return nil, apierror.ErrInvalidRequest
	}
	if err != nil {
		logger.Error("Ошибка mock-оплаты orderID=%d: %v", orderID, err)
		return nil, err
	}
	status := "failed"
	if success {
		status = "success"
		logger.Info("Mock-оплата успешно завершена: orderID=%d paymentID=%d", orderID, paymentID)
	} else {
		logger.Warn("Mock-оплата отклонена: orderID=%d paymentID=%d", orderID, paymentID)
	}
	return &payments.PayOutput{
		OrderID:      orderID,
		PaymentID:    paymentID,
		Status:       status,
		Amount:       data.PayableAmount,
		BonusSpent:   data.BonusSpent,
		BonusEarned:  data.BonusEarned,
		BonusBalance: balance,
	}, nil
}

func mockSuccess() (bool, error) {
	var value [1]byte
	if _, err := rand.Read(value[:]); err != nil {
		return false, err
	}
	return int(value[0])%100 >= 3, nil
}
