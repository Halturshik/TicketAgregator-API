package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
)

type transaction struct {
	tx *sql.Tx
}

func (r *Repository) WithinTransaction(ctx context.Context, operation func(checkout.Transaction) error) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin checkout tx: %w", err)
	}
	defer tx.Rollback()

	if err := operation(&transaction{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit checkout tx: %w", err)
	}
	return nil
}
