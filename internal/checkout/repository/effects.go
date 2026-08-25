package repository

import (
	"context"
	"fmt"
)

func (t *transaction) LockUserEffects(ctx context.Context, userID int) error {
	if _, err := t.tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1::BIGINT)`, userID); err != nil {
		return fmt.Errorf("lock payment effects: %w", err)
	}
	return nil
}
