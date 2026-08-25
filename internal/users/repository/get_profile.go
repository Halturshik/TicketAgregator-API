package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

func (r *Repository) GetProfile(ctx context.Context, userID int) (*users.Profile, error) {
	var p users.Profile
	var birthDate time.Time

	err := r.DB.QueryRowContext(ctx, `
		SELECT id, email, first_name, COALESCE(middle_name, ''), last_name, birth_date, is_russian, bonus_points
		FROM users
		WHERE id = $1
	`, userID).Scan(&p.ID, &p.Email, &p.FirstName, &p.MiddleName, &p.LastName, &birthDate, &p.IsRussian, &p.BonusPoints)
	if err == sql.ErrNoRows {
		return nil, users.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	p.BirthDate = birthDate.Format("2006-01-02")
	return &p, nil
}
