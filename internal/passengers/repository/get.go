package repository

import (
	"context"
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func (r *Repository) GetOwned(ctx context.Context, ownerUserID int, passengerID int) (*passengers.Passenger, error) {
	item, err := scanPassenger(r.DB.QueryRowContext(ctx, `
		SELECT id, first_name, COALESCE(middle_name, ''), last_name, birth_date, is_russian
		FROM saved_passengers
		WHERE id = $1 AND owner_user_id = $2 AND deleted_at IS NULL
	`, passengerID, ownerUserID))
	if err == sql.ErrNoRows {
		return nil, passengers.ErrNotFound
	}
	return item, err
}
