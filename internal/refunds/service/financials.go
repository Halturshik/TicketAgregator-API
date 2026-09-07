package service

import (
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

type refundFinancials struct {
	currentTotalPrice int
	bonusSpent        int
	bonusEarned       int
	payableAmount     int
	cashAmount        int
	bonusRestored     int
	bonusRevoked      int
}

func calculateRefundFinancials(
	order *refunds.OrderData,
	refundedTicketTotal int,
	supplierRefundAmount int,
) (*refundFinancials, error) {
	if order == nil {
		return nil, invalidFinancialState("order is nil")
	}
	if refundedTicketTotal <= 0 || refundedTicketTotal > order.CurrentTotalPrice {
		return nil, invalidFinancialState(
			"refunded ticket total %d exceeds current total %d",
			refundedTicketTotal, order.CurrentTotalPrice,
		)
	}
	if supplierRefundAmount < 0 || supplierRefundAmount > refundedTicketTotal {
		return nil, refunds.ErrSupplierMismatch
	}
	if order.BonusSpent < 0 || order.BonusSpent > bonus.MaxSpend(order.CurrentTotalPrice) ||
		order.BonusEarned < 0 || order.PayableAmount < 0 {
		return nil, invalidFinancialState(
			"order %d has current=%d spent=%d earned=%d payable=%d",
			order.ID, order.CurrentTotalPrice, order.BonusSpent, order.BonusEarned, order.PayableAmount,
		)
	}

	expectedEarned := 0
	if order.UserID != nil {
		expectedEarned = bonus.Earned(order.CurrentTotalPrice)
	}
	if order.BonusEarned != expectedEarned {
		return nil, invalidFinancialState(
			"order %d has earned=%d, expected=%d",
			order.ID, order.BonusEarned, expectedEarned,
		)
	}

	currentTotalPrice := order.CurrentTotalPrice - refundedTicketTotal
	bonusSpent := min(order.BonusSpent, bonus.MaxSpend(currentTotalPrice))
	bonusEarned := 0
	if order.UserID != nil {
		bonusEarned = bonus.Earned(currentTotalPrice)
	}
	bonusRestored := order.BonusSpent - bonusSpent
	bonusRevoked := order.BonusEarned - bonusEarned
	cashAmount := supplierRefundAmount - bonusRestored
	if cashAmount < 0 || cashAmount > order.PayableAmount {
		return nil, invalidFinancialState(
			"order %d produces cash refund=%d with payable=%d",
			order.ID, cashAmount, order.PayableAmount,
		)
	}

	return &refundFinancials{
		currentTotalPrice: currentTotalPrice,
		bonusSpent:        bonusSpent,
		bonusEarned:       bonusEarned,
		payableAmount:     order.PayableAmount - cashAmount,
		cashAmount:        cashAmount,
		bonusRestored:     bonusRestored,
		bonusRevoked:      bonusRevoked,
	}, nil
}

func invalidFinancialState(format string, args ...any) error {
	return fmt.Errorf("%w: %s", refunds.ErrInvalidFinancialState, fmt.Sprintf(format, args...))
}
