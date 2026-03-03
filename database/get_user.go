package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/database/model"
)

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User

	err := s.DB.QueryRowContext(ctx,
		`SELECT id, first_name, last_name, email, password_hash, is_russian, birth_date
		 FROM users WHERE email = $1`,
		email,
	).Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.PasswordHash,
		&u.IsRussian,
		&u.BirthDate,
	)

	if err == sql.ErrNoRows {
		return nil, errs.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}
