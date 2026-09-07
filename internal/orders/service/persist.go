package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/orders/number"
)

const orderNumberGenerationAttempts = 5

func (s *Service) persistOrder(ctx context.Context, params orders.CreateOrderParams) (*orders.Order, error) {
	for range orderNumberGenerationAttempts {
		orderNumber, err := number.NewOrderNumber()
		if err != nil {
			return nil, fmt.Errorf("generate order number: %w", err)
		}
		params.OrderNumber = orderNumber
		order, err := s.repo.Create(ctx, params)
		if !errors.Is(err, orders.ErrOrderNumberConflict) {
			return order, err
		}
	}
	return nil, fmt.Errorf("generate unique order number: %w", orders.ErrOrderNumberConflict)
}
