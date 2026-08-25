package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func (r *Repository) Update(
	ctx context.Context,
	ownerUserID int,
	passengerID int,
	in passengers.SavePassengerInput,
	birthDate time.Time,
) (*passengers.Passenger, error) {
	item, err := scanPassenger(r.DB.QueryRowContext(ctx, `
		UPDATE saved_passengers
		SET first_name = $1, middle_name = $2, last_name = $3,
			birth_date = $4, is_russian = $5, updated_at = NOW()
		WHERE id = $6 AND owner_user_id = $7 AND deleted_at IS NULL
		RETURNING id, first_name, COALESCE(middle_name, ''), last_name, birth_date, is_russian
	`, in.FirstName, in.MiddleName, in.LastName, birthDate, in.IsRussian, passengerID, ownerUserID))
	if err == sql.ErrNoRows {
		return nil, passengers.ErrNotFound
	}
	return item, err
}
