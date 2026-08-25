package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

func (r *Repository) UpdateProfile(ctx context.Context, userID int, in users.UpdateProfileInput, birthDate time.Time) (*users.Profile, error) {
	var p users.Profile
	var storedBirthDate time.Time

	err := r.DB.QueryRowContext(ctx, `
		UPDATE users
		SET first_name = $1,
		    middle_name = $2,
		    last_name = $3,
		    birth_date = $4,
		    is_russian = $5,
		    updated_at = NOW()
		WHERE id = $6
		RETURNING id, email, first_name, COALESCE(middle_name, ''), last_name, birth_date, is_russian, bonus_points
	`, in.FirstName, in.MiddleName, in.LastName, birthDate, in.IsRussian, userID).
		Scan(&p.ID, &p.Email, &p.FirstName, &p.MiddleName, &p.LastName, &storedBirthDate, &p.IsRussian, &p.BonusPoints)
	if err == sql.ErrNoRows {
		return nil, users.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}

	p.BirthDate = storedBirthDate.Format("2006-01-02")
	return &p, nil
}
