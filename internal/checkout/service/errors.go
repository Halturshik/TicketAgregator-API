package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func mapPaymentError(ctx context.Context, orderID int, err error) error {
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
		slog.WarnContext(ctx, "Попытка оплаты заказа с истёкшим сроком", slog.Int("order_id", orderID))
		return apierror.ErrOrderExpired
	default:
		return fmt.Errorf("pay order %d: %w", orderID, err)
	}
}
