package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
)

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*auth.UserAuth, error) {
	var u auth.UserAuth

	err := r.DB.QueryRowContext(ctx,
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
		return nil, auth.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}
