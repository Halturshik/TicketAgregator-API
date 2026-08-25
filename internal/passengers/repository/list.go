package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func (r *Repository) List(ctx context.Context, ownerUserID int) ([]passengers.Passenger, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, first_name, COALESCE(middle_name, ''), last_name, birth_date, is_russian
		FROM saved_passengers
		WHERE owner_user_id = $1 AND deleted_at IS NULL
		ORDER BY id DESC
	`, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("list passengers: %w", err)
	}
	defer rows.Close()

	result := make([]passengers.Passenger, 0)
	for rows.Next() {
		passenger, err := scanPassenger(rows)
		if err != nil {
			return nil, fmt.Errorf("scan passenger: %w", err)
		}
		result = append(result, *passenger)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate passengers: %w", err)
	}
	return result, nil
}
