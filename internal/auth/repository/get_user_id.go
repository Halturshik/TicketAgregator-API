package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
)

func (r *Repository) GetUserByID(ctx context.Context, id int) (*auth.UserAuth, error) {
	var u auth.UserAuth

	err := r.DB.QueryRowContext(ctx,
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
		return nil, auth.ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &u, nil
}
