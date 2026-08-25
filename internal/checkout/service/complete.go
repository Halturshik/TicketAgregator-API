package service

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func (s *Service) completePayment(
	ctx context.Context,
	tx checkout.Transaction,
	order orders.PaymentData,
) (*int, error) {
	if err := tx.MarkPaid(ctx, order.ID); err != nil {
		return nil, err
	}
	if order.UserID == nil {
		return nil, nil
	}
	userID := *order.UserID
	if err := tx.LockUserEffects(ctx, userID); err != nil {
		return nil, err
	}
	if err := applyBonusChanges(ctx, tx, userID, order); err != nil {
		return nil, err
	}
	items, err := tx.ListPassengersForPayment(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	if err := s.passengers.SyncAfterPayment(ctx, tx, tx, userID, items); err != nil {
		return nil, fmt.Errorf("sync saved passengers: %w", err)
	}
	balance, err := tx.GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func applyBonusChanges(ctx context.Context, tx checkout.Transaction, userID int, order orders.PaymentData) error {
	if order.BonusSpent > 0 {
		if err := tx.Spend(ctx, userID, order.ID, order.BonusSpent); err != nil {
			return err
		}
	}
	if order.BonusEarned > 0 {
		if err := tx.Earn(ctx, userID, order.ID, order.BonusEarned); err != nil {
			return err
		}
	}
	return nil
}
