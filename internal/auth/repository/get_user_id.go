package repository

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Repository) GetUserByID(ctx context.Context, id int) (*UserAuth, error) {
	var u UserAuth

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
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &u, nil
}
