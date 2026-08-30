package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/lib/pq"
)

func (r *Repository) CreateOrGet(ctx context.Context, operation supplier.RefundOperation) (*supplier.RefundOperation, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, fmt.Errorf("begin supplier refund tx: %w", err)
	}
	defer tx.Rollback()

	var insertedID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO supplier_simulator.refund_operations
			(id, provider_code, idempotency_key, request_hash, status, failure_code)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
		ON CONFLICT (provider_code, idempotency_key) DO NOTHING
		RETURNING id
	`, operation.ID, operation.ProviderCode, operation.IdempotencyKey,
		operation.RequestHash, operation.Status, operation.FailureCode).Scan(&insertedID)
	if err == sql.ErrNoRows {
		existing, err := loadByKey(ctx, tx, operation.ProviderCode, operation.IdempotencyKey)
		if err != nil {
			return nil, false, err
		}
		if err := tx.Commit(); err != nil {
			return nil, false, fmt.Errorf("commit existing supplier refund tx: %w", err)
		}
		return existing, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("insert supplier refund operation: %w", err)
	}

	if err := ensureTicketsNotRefunded(ctx, tx, operation); err != nil {
		return nil, false, err
	}
	for _, item := range operation.Items {
		policy, err := json.Marshal(item.Policy)
		if err != nil {
			return nil, false, fmt.Errorf("marshal supplier refund policy: %w", err)
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO supplier_simulator.refund_items
				(operation_id, provider_code, aggregator_ticket_id, ticket_number,
				 supplier_offer_id, fare_type, departure_at, gross_amount, refunded,
				 reason, refund_percent, refund_amount, refund_policy)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`, operation.ID, operation.ProviderCode, item.TicketID, item.TicketNumber,
			item.SupplierOfferID, item.FareType, time.Unix(item.DepartureUnix, 0).UTC(),
			item.GrossAmount, item.Refunded, item.Reason, item.RefundPercent,
			item.RefundAmount, policy)
		if err != nil {
			if constraintViolation(err, "supplier_refund_items_refunded_ticket_unique") {
				return nil, false, supplier.ErrTicketAlreadyRefunded
			}
			return nil, false, fmt.Errorf("insert supplier refund item: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit supplier refund tx: %w", err)
	}
	return &operation, true, nil
}

func ensureTicketsNotRefunded(ctx context.Context, tx *sql.Tx, operation supplier.RefundOperation) error {
	for _, item := range operation.Items {
		var exists bool
		if err := tx.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM supplier_simulator.refund_items
				WHERE provider_code = $1 AND supplier_offer_id = $2
					AND ticket_number = $3 AND refunded
			)
		`, operation.ProviderCode, item.SupplierOfferID, item.TicketNumber).Scan(&exists); err != nil {
			return fmt.Errorf("check supplier refunded ticket: %w", err)
		}
		if exists {
			return supplier.ErrTicketAlreadyRefunded
		}
	}
	return nil
}

func loadByKey(ctx context.Context, tx *sql.Tx, providerCode string, idempotencyKey string) (*supplier.RefundOperation, error) {
	var operation supplier.RefundOperation
	err := tx.QueryRowContext(ctx, `
		SELECT id, provider_code, idempotency_key, request_hash, status,
			COALESCE(failure_code, '')
		FROM supplier_simulator.refund_operations
		WHERE provider_code = $1 AND idempotency_key = $2
	`, providerCode, idempotencyKey).Scan(
		&operation.ID, &operation.ProviderCode, &operation.IdempotencyKey,
		&operation.RequestHash, &operation.Status, &operation.FailureCode,
	)
	if err != nil {
		return nil, fmt.Errorf("load supplier refund operation: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT aggregator_ticket_id, ticket_number, supplier_offer_id, fare_type,
			departure_at, gross_amount, refunded, reason, refund_percent,
			refund_amount, refund_policy
		FROM supplier_simulator.refund_items
		WHERE operation_id = $1
		ORDER BY aggregator_ticket_id
	`, operation.ID)
	if err != nil {
		return nil, fmt.Errorf("load supplier refund items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item supplier.RefundOperationItem
		var departure time.Time
		var policy []byte
		if err := rows.Scan(
			&item.TicketID, &item.TicketNumber, &item.SupplierOfferID, &item.FareType,
			&departure, &item.GrossAmount, &item.Refunded, &item.Reason,
			&item.RefundPercent, &item.RefundAmount, &policy,
		); err != nil {
			return nil, fmt.Errorf("scan supplier refund item: %w", err)
		}
		item.DepartureUnix = departure.Unix()
		if err := json.Unmarshal(policy, &item.Policy); err != nil {
			return nil, fmt.Errorf("decode supplier refund policy: %w", err)
		}
		operation.Items = append(operation.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate supplier refund items: %w", err)
	}
	return &operation, nil
}

func constraintViolation(err error, name string) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Constraint == name
}
