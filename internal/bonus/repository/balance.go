package repository

import (
	"context"
	"fmt"
)

func (r *Repository) GetBalance(ctx context.Context, userID int) (int, error) {
	var balance int
	if err := r.DB.QueryRowContext(ctx, `SELECT bonus_points FROM users WHERE id = $1`, userID).Scan(&balance); err != nil {
		return 0, fmt.Errorf("get bonus balance: %w", err)
	}
	return balance, nil
}
