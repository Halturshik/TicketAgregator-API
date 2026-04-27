package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/database/model"
)

func (s *Store) GetUserByID(ctx context.Context, id int) (*model.UserAuth, error) {
	var u model.UserAuth

	err := s.DB.QueryRowContext(ctx,
		`SELECT id, email, password_hash, token_version
		 FROM users WHERE id = $1`,
		id,
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
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &u, nil
}
