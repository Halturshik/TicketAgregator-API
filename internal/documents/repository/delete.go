package repository

import "context"

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
		return err
	}
	return requireSingleRow(result)
}
