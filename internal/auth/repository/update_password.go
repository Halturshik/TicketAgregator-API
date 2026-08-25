package repository

import "context"

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

	return newVersion, err
}
