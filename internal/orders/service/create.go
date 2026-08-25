package service

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

func (s *Service) Create(ctx context.Context, userID *int, in orders.CreateOrderInput) (*orders.Order, error) {
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}

	trip, err := s.loadTrip(ctx, in)
	if err != nil {
		return nil, err
	}
	passengers, tickets, err := s.buildOrderContent(ctx, userID, trip.directions, in.Passengers)
	if err != nil {
		return nil, err
	}
	pricing, err := s.calculatePricing(ctx, userID, in.UseBonus, trip.total)
	if err != nil {
		return nil, err
	}
	guestEmail, guestPaymentToken, err := guestPaymentData(userID, in.GuestEmail)
	if err != nil {
		return nil, err
	}

	order, err := s.repo.Create(ctx, orders.CreateOrderParams{
		UserID:            userID,
		GuestEmail:        guestEmail,
		GuestPaymentToken: guestPaymentToken,
		TotalPrice:        trip.total,
		BonusSpent:        pricing.spent,
		BonusEarned:       pricing.earned,
		PayableAmount:     pricing.payable,
		ExpiresAt:         s.now().UTC().Add(orders.OrderTTL),
		Passengers:        passengers,
		Tickets:           tickets,
	})
	if err != nil {
		logger.Error("Ошибка создания заказа: %v", err)
		return nil, err
	}
	logger.Info(
		"Заказ создан: orderID=%d tickets=%d passengers=%d total=%d payable=%d",
		order.ID, len(order.Tickets), len(passengers), order.TotalPrice, order.PayableAmount,
	)
	return order, nil
}

func validateCreateInput(in orders.CreateOrderInput) error {
	if in.SearchID == "" || in.TripOptionID == "" || len(in.Passengers) == 0 || in.UseBonus < 0 {
		return apierror.ErrInvalidRequest
	}
	return nil
}
