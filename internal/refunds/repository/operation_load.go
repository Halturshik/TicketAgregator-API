package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

type operationQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func loadOperation(ctx context.Context, db operationQueryer, query string, arg any, lockItems bool) (*refunds.Operation, error) {
	var operation refunds.Operation
	var userID sql.NullInt64
	var bonusBalance sql.NullInt64
	var bonusDebt sql.NullInt64
	err := db.QueryRowContext(ctx, query, arg).Scan(
		&operation.ID, &operation.PaymentID, &operation.Status, &operation.IdempotencyKey,
		&operation.RequestHash, &operation.SupplierCode, &operation.SupplierRefundID,
		&operation.FailureCode, &operation.LastError, &operation.AttemptCount,
		&operation.NextRetryAt, &operation.ReconciliationDeadline,
		&operation.CashAmount, &operation.BonusRestored, &operation.BonusRevoked,
		&bonusBalance, &bonusDebt, &operation.Order.ID, &userID,
		&operation.Order.GuestToken, &operation.Order.Status,
	)
	if err == sql.ErrNoRows {
		return nil, refunds.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load refund operation: %w", err)
	}
	if userID.Valid {
		value := int(userID.Int64)
		operation.Order.UserID = &value
	}
	if bonusBalance.Valid {
		value := int(bonusBalance.Int64)
		operation.BonusBalanceAfter = &value
	}
	if bonusDebt.Valid {
		value := int(bonusDebt.Int64)
		operation.BonusDebtAfter = &value
	}

	itemQuery := `
		SELECT ri.ticket_id, ri.ticket_number, t.status, ri.supplier_code,
			ri.supplier_offer_id::text, ri.fare_type, t.refund_policy_version,
			t.refund_policy_snapshot, ri.gross_amount, ri.bonus_restored,
			ri.bonus_revoked, t.payable_amount, ri.departure_at, ri.reason,
			ri.refund_percent, ri.cash_refunded
		FROM refund_items ri
		JOIN tickets t ON t.id = ri.ticket_id
		WHERE ri.refund_id = $1
		ORDER BY ri.ticket_id`
	if lockItems {
		itemQuery += " FOR UPDATE OF ri, t"
	}
	rows, err := db.QueryContext(ctx, itemQuery, operation.ID)
	if err != nil {
		return nil, fmt.Errorf("load refund operation items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item refunds.OperationItem
		var policy []byte
		if err := rows.Scan(
			&item.ID, &item.TicketNumber, &item.Status, &item.SupplierCode,
			&item.SupplierOfferID, &item.FareType, &item.RefundPolicyVersion,
			&policy, &item.GrossAmount, &item.BonusRestored, &item.BonusRevoked,
			&item.PayableAmount, &item.DepartureAt, &item.Reason,
			&item.RefundPercent, &item.CashRefunded,
		); err != nil {
			return nil, fmt.Errorf("scan refund operation item: %w", err)
		}
		item.Price = item.GrossAmount
		item.BonusSpent = item.BonusRestored
		item.BonusEarned = item.BonusRevoked
		if err := json.Unmarshal(policy, &item.RefundPolicy); err != nil {
			return nil, fmt.Errorf("decode refund operation policy: %w", err)
		}
		operation.Items = append(operation.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate refund operation items: %w", err)
	}
	return &operation, nil
}
