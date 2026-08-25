package service

import (
	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func authorizePayment(order orders.PaymentData, expectedUserID *int, guestPaymentToken string) error {
	if order.UserID != nil {
		if expectedUserID == nil || *order.UserID != *expectedUserID {
			return checkout.ErrPaymentForbidden
		}
		return nil
	}
	if order.GuestToken == "" || guestPaymentToken == "" || order.GuestToken != guestPaymentToken {
		return checkout.ErrPaymentForbidden
	}
	return nil
}

func ensureOrderPayable(order orders.PaymentData) error {
	if order.Status != orders.OrderStatusCreated {
		return orders.ErrCannotBePaid
	}
	return nil
}
