package repository

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

const operationColumns = `
	SELECT r.id, r.payment_id, r.status, r.idempotency_key::text, r.request_hash,
		r.supplier_code, COALESCE(r.supplier_refund_id::text, ''),
		COALESCE(r.failure_code, ''), COALESCE(r.last_error, ''), r.attempt_count,
		r.next_retry_at, r.reconciliation_deadline,
		r.cash_amount, r.bonus_restored, r.bonus_revoked,
		r.bonus_balance_after, r.bonus_debt_after,
		o.id, o.user_id, COALESCE(o.guest_payment_token::text, ''), o.status,
		o.current_total_price, o.bonus_spent, o.bonus_earned, o.payable_amount
	FROM refunds r
	JOIN orders o ON o.id = r.order_id
`

func (r *Repository) GetByKey(ctx context.Context, idempotencyKey string) (*refunds.Operation, error) {
	return loadOperation(ctx, r.db, operationColumns+" WHERE r.idempotency_key = $1", idempotencyKey, false)
}

func (t *transaction) LockOperation(ctx context.Context, refundID int) (*refunds.Operation, error) {
	return loadOperation(ctx, t.tx, operationColumns+" WHERE r.id = $1 FOR UPDATE OF r", refundID, true)
}

func (t *transaction) GetByKey(ctx context.Context, idempotencyKey string) (*refunds.Operation, error) {
	return loadOperation(ctx, t.tx, operationColumns+" WHERE r.idempotency_key = $1", idempotencyKey, false)
}
