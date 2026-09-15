package repository

import (
	"context"
	"fmt"
)

func (r *Repository) IsEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email existence: %w", err)
	}

	return exists, nil
}
