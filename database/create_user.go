package database

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/database/model"
	"github.com/lib/pq"
)

func (s *Store) CreateUser(ctx context.Context, u model.CreateUserParams) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
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
			return 0, errs.ErrDuplicateEmail
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return id, nil
}
