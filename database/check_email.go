package database

import "context"

func (s *Store) IsEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := s.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)`, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
