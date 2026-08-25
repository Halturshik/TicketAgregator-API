package service

import (
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func mapPaymentError(orderID int, err error) error {
	switch {
	case errors.Is(err, orders.ErrNotFound):
		return apierror.ErrNotFound
	case errors.Is(err, checkout.ErrPaymentForbidden):
		return apierror.ErrForbidden
	case errors.Is(err, bonus.ErrInsufficientBalance):
		return apierror.ErrInsufficientBonus
	case errors.Is(err, orders.ErrCannotBePaid):
		return apierror.ErrInvalidRequest
	case errors.Is(err, orders.ErrExpired):
		logger.Warn("Попытка оплаты заказа с истёкшим сроком: orderID=%d", orderID)
		return apierror.ErrOrderExpired
	default:
		logger.Error("Ошибка mock-оплаты orderID=%d: %v", orderID, err)
		return err
	}
}
