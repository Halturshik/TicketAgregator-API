package repository

import "context"

func (r *Repository) GetBalance(ctx context.Context, userID int) (int, error) {
	var balance int
	err := r.DB.QueryRowContext(ctx, `SELECT bonus_points FROM users WHERE id = $1`, userID).Scan(&balance)
	return balance, err
}
