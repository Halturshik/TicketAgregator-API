package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

type transaction struct {
	tx *sql.Tx
}

func (r *Repository) WithinTransaction(ctx context.Context, operation func(refunds.Transaction) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin refund tx: %w", err)
	}
	defer tx.Rollback()

	if err := operation(&transaction{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit refund tx: %w", err)
	}
	return nil
}
