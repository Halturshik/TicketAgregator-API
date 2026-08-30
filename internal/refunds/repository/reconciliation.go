package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

func (r *Repository) ListProcessing(ctx context.Context, now time.Time, maxAttempts int, limit int) ([]refunds.Operation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id
		FROM refunds
		WHERE status = $1 AND next_retry_at <= $2
			AND reconciliation_deadline > $2 AND attempt_count < $3
		ORDER BY next_retry_at, id
		LIMIT $4
	`, refunds.StatusProcessing, now, maxAttempts, limit)
	if err != nil {
		return nil, fmt.Errorf("list processing refunds: %w", err)
	}
	defer rows.Close()
	ids := make([]int, 0, limit)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan processing refund id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate processing refunds: %w", err)
	}
	operations := make([]refunds.Operation, 0, len(ids))
	for _, id := range ids {
		operation, err := loadOperation(ctx, r.db, operationColumns+" WHERE r.id = $1", id, false)
		if err != nil {
			return nil, err
		}
		operations = append(operations, *operation)
	}
	return operations, nil
}

func (r *Repository) MarkReviewDue(ctx context.Context, now time.Time, maxAttempts int) (int, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE refunds
		SET status = $1, failure_code = $2, updated_at = NOW()
		WHERE status = $3 AND (attempt_count >= $4 OR reconciliation_deadline <= $5)
	`, refunds.StatusRequiresReview, refunds.FailureReconciliationExhausted,
		refunds.StatusProcessing, maxAttempts, now)
	if err != nil {
		return 0, fmt.Errorf("mark refunds requiring review: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count refunds requiring review: %w", err)
	}
	return int(affected), nil
}

func (r *Repository) RecordAttemptError(ctx context.Context, params refunds.AttemptErrorParams) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE refunds
		SET attempt_count = attempt_count + 1, last_error = LEFT($1, 500),
			status = $2, failure_code = NULLIF($3, ''), next_retry_at = $4,
			updated_at = NOW()
		WHERE id = $5 AND status = $6
	`, params.Message, params.Status, params.FailureCode, params.NextRetryAt,
		params.RefundID, refunds.StatusProcessing)
	if err != nil {
		return fmt.Errorf("record refund attempt error: %w", err)
	}
	if err := requireAffectedRows(result, 1); err != nil {
		return fmt.Errorf("record refund attempt error: %w", err)
	}
	return nil
}
