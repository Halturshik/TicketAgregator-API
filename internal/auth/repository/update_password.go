package repository

import (
	"context"
	"fmt"
)

func (r *Repository) UpdatePassword(ctx context.Context, userID int, hash string) (int, error) {
	var newVersion int

	err := r.DB.QueryRowContext(ctx,
		`UPDATE users 
		 SET password_hash = $1,
		     token_version = token_version + 1
		 WHERE id = $2
		 RETURNING token_version`,
		hash,
		userID,
	).Scan(&newVersion)

	if err != nil {
		return 0, fmt.Errorf("update password: %w", err)
	}
	return newVersion, nil
}
