package repository

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Repository) GetUserByEmail(ctx context.Context, email string) (*UserAuth, error) {
	var u UserAuth

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
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}
