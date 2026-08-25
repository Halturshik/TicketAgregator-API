package repository

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func (r *Repository) Delete(ctx context.Context, ownerUserID int, passengerID int) error {
	result, err := r.DB.ExecContext(ctx, `
		UPDATE saved_passengers
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND owner_user_id = $2 AND deleted_at IS NULL
	`, passengerID, ownerUserID)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected != 1 {
		return passengers.ErrNotFound
	}
	return nil
}
