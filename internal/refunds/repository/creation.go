package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func (t *transaction) SuccessfulPaymentID(ctx context.Context, orderID int) (int, error) {
	var paymentID int
	err := t.tx.QueryRowContext(ctx, `
		SELECT id
		FROM payments
		WHERE order_id = $1 AND status = 'success'
	`, orderID).Scan(&paymentID)
	if err == sql.ErrNoRows {
		return 0, refunds.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("load successful payment: %w", err)
	}
	return paymentID, nil
}

func (t *transaction) CreateProcessing(ctx context.Context, params refunds.CreateParams) (int, bool, error) {
	var refundID int
	err := t.tx.QueryRowContext(ctx, `
		INSERT INTO refunds
			(order_id, payment_id, status, idempotency_key, request_hash, supplier_code,
			 cash_amount, bonus_restored, bonus_revoked, next_retry_at, reconciliation_deadline)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING id
	`, params.OrderID, params.PaymentID, refunds.StatusProcessing, params.IdempotencyKey,
		params.RequestHash, params.SupplierCode, params.CashAmount,
		params.BonusRestored, params.BonusRevoked, params.NextRetryAt,
		params.ReconciliationDeadline).Scan(&refundID)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("insert processing refund: %w", err)
	}

	for _, item := range params.Items {
		result, err := t.tx.ExecContext(ctx, `
			INSERT INTO refund_items
				(refund_id, ticket_id, supplier_code, ticket_number, supplier_offer_id,
				 fare_type, departure_at, reason, refund_percent, gross_amount,
				 supplier_refund_amount)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, refundID, item.TicketID, item.SupplierCode, item.TicketNumber,
			item.SupplierOfferID, item.FareType, item.DepartureAt, item.Reason,
			item.RefundPercent, item.GrossAmount, item.SupplierRefundAmount)
		if err != nil {
			return 0, false, fmt.Errorf("insert processing refund item: %w", err)
		}
		if err := requireAffectedRows(result, 1); err != nil {
			return 0, false, fmt.Errorf("insert processing refund item: %w", err)
		}
	}
	return refundID, true, nil
}
