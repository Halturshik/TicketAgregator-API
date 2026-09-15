package repository

import (
	"context"
	"fmt"
)

func (r *Repository) Delete(ctx context.Context, ownerUserID int, documentID int) error {
	result, err := r.DB.ExecContext(ctx, `
		DELETE FROM documents d
		WHERE d.id = $1 AND (
			d.owner_user_id = $2
			OR EXISTS (
				SELECT 1 FROM saved_passengers p
				WHERE p.id = d.passenger_id AND p.owner_user_id = $2 AND p.deleted_at IS NULL
			)
		)
	`, documentID, ownerUserID)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	if err := requireSingleRow(result); err != nil {
		return fmt.Errorf("delete document result: %w", err)
	}
	return nil
}
