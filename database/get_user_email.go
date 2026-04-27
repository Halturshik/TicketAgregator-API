package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/database/model"
)

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.UserAuth, error) {
	var u model.UserAuth

	err := s.DB.QueryRowContext(ctx,
		`SELECT id, email, password_hash, token_version
		 FROM users WHERE email = $1`,
		email,
	).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.TokenVersion,
	)

	if err == sql.ErrNoRows {
		return nil, errs.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}
