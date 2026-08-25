package repository

import "context"

func (r *Repository) IsEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
