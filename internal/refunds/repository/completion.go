package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func (t *transaction) MarkSuccess(ctx context.Context, params refunds.SuccessParams) error {
	for _, item := range params.Items {
		result, err := t.tx.ExecContext(ctx, `
			UPDATE refund_items
			SET reason = $1, refund_percent = $2, supplier_refund_amount = $3
			WHERE refund_id = $4 AND ticket_id = $5
		`, item.Reason, item.RefundPercent, item.SupplierRefundAmount,
			params.RefundID, item.TicketID)
		if err != nil {
			return fmt.Errorf("update successful refund item: %w", err)
		}
		if err := requireAffectedRows(result, 1); err != nil {
			return fmt.Errorf("update successful refund item: %w", err)
		}
	}
	result, err := t.tx.ExecContext(ctx, `
		UPDATE refunds
		SET status = $1, supplier_refund_id = $2, cash_amount = $3,
			bonus_restored = $4, bonus_revoked = $5,
			bonus_balance_after = $6, bonus_debt_after = $7,
			attempt_count = attempt_count + 1, failure_code = NULL,
			last_error = NULL, updated_at = NOW(), completed_at = NOW()
		WHERE id = $8 AND status = $9
	`, refunds.StatusSuccess, params.SupplierRefundID, params.CashAmount,
		params.BonusRestored, params.BonusRevoked, params.BonusBalanceAfter,
		params.BonusDebtAfter, params.RefundID, refunds.StatusProcessing)
	if err != nil {
		return fmt.Errorf("complete refund: %w", err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return fmt.Errorf("complete refund: %w", err)
	}
	return nil
}

func (t *transaction) MarkFailed(ctx context.Context, params refunds.FailureParams) error {
	for _, item := range params.Items {
		result, err := t.tx.ExecContext(ctx, `
			UPDATE refund_items
			SET reason = $1, refund_percent = 0, supplier_refund_amount = 0
			WHERE refund_id = $2 AND ticket_id = $3
		`, item.Reason, params.RefundID, item.TicketID)
		if err != nil {
			return fmt.Errorf("update failed refund item: %w", err)
		}
		if err := requireAffectedRows(result, 1); err != nil {
			return fmt.Errorf("update failed refund item: %w", err)
		}
	}
	result, err := t.tx.ExecContext(ctx, `
		UPDATE refunds
		SET status = $1, supplier_refund_id = $2,
			cash_amount = 0, bonus_restored = 0,
			bonus_revoked = 0, bonus_debt_created = 0,
			bonus_balance_after = NULL, bonus_debt_after = NULL,
			attempt_count = attempt_count + 1, failure_code = $3,
			last_error = NULL, updated_at = NOW(), completed_at = NOW()
		WHERE id = $4 AND status = $5
	`, refunds.StatusFailed, params.SupplierRefundID, params.FailureCode,
		params.RefundID, refunds.StatusProcessing)
	if err != nil {
		return fmt.Errorf("fail refund: %w", err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return fmt.Errorf("fail refund: %w", err)
	}
	return nil
}

func (t *transaction) CountRefundedTickets(ctx context.Context, orderID int) (int, int, error) {
	var total int
	var refunded int
	if err := t.tx.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status = $2)
		FROM tickets
		WHERE order_id = $1
	`, orderID, orders.TicketStatusRefunded).Scan(&total, &refunded); err != nil {
		return 0, 0, fmt.Errorf("count refunded tickets: %w", err)
	}
	return total, refunded, nil
}

func (t *transaction) UpdateOrderAfterRefund(ctx context.Context, params refunds.OrderRefundParams) error {
	result, err := t.tx.ExecContext(ctx, `
		UPDATE orders
		SET status = $1, current_total_price = $2, bonus_spent = $3,
			bonus_earned = $4, payable_amount = $5
		WHERE id = $6 AND status IN ($7, $8)
	`, params.Status, params.CurrentTotalPrice, params.BonusSpent, params.BonusEarned,
		params.PayableAmount, params.OrderID, orders.OrderStatusPaid, orders.OrderStatusPartiallyRefunded)
	if err != nil {
		return fmt.Errorf("update refunded order status: %w", err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return fmt.Errorf("update refunded order status: %w", err)
	}
	return nil
}
