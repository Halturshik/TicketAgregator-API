package service

import (
	"context"
	"log/slog"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (s *Service) Create(ctx context.Context, userID *int, in orders.CreateOrderInput) (*orders.Order, error) {
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}

	trip, err := s.loadTrip(ctx, in)
	if err != nil {
		return nil, err
	}
	passengers, tickets, err := s.buildOrderContent(ctx, userID, trip, in.Passengers)
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

	order, err := s.persistOrder(ctx, orders.CreateOrderParams{
		UserID:            userID,
		GuestEmail:        guestEmail,
		GuestPaymentToken: guestPaymentToken,
		TotalPrice:        trip.total,
		CurrentTotalPrice: trip.total,
		BonusSpent:        pricing.spent,
		BonusEarned:       pricing.earned,
		PayableAmount:     pricing.payable,
		ExpiresAt:         s.now().UTC().Add(orders.OrderTTL),
		Passengers:        passengers,
		Tickets:           tickets,
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Заказ создан",
		slog.Int("order_id", order.ID),
		slog.String("order_number", order.OrderNumber),
		slog.Int("tickets", len(order.Tickets)),
		slog.Int("passengers", len(passengers)),
		slog.Int("total_price", order.TotalPrice),
		slog.Int("payable_amount", order.PayableAmount),
	)
	return order, nil
}

func validateCreateInput(in orders.CreateOrderInput) error {
	if in.SearchID == "" || in.TripOptionID == "" || len(in.Passengers) == 0 || in.UseBonus < 0 {
		return apierror.ErrInvalidRequest
	}
	return nil
}
