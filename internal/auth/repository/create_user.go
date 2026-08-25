package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/lib/pq"
)

func (r *Repository) CreateUser(ctx context.Context, u auth.CreateUserParams) (int64, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var id int64
	query := `
		INSERT INTO users 
		(first_name, middle_name, last_name, birth_date, email, password_hash, is_russian)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
	`

	err = tx.QueryRowContext(ctx, query, u.FirstName, u.MiddleName, u.LastName, u.BirthDate, u.Email, u.PasswordHash, u.IsRussian).Scan(&id)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return 0, auth.ErrDuplicateEmail
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return id, nil
}
