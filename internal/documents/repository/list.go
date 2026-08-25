package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func (r *Repository) ListForUser(ctx context.Context, ownerUserID int) ([]documents.Document, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, owner_user_id, passenger_id, type, number, verification_status,
			document_fingerprint, expires_at, last_checked_at
		FROM documents
		WHERE owner_user_id = $1
		ORDER BY id DESC
	`, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("list user documents: %w", err)
	}
	defer rows.Close()
	return collect(rows)
}

func (r *Repository) ListForPassenger(ctx context.Context, ownerUserID int, passengerID int) ([]documents.Document, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT d.id, d.owner_user_id, d.passenger_id, d.type, d.number, d.verification_status,
			d.document_fingerprint, d.expires_at, d.last_checked_at
		FROM documents d
		JOIN saved_passengers p ON p.id = d.passenger_id
		WHERE d.passenger_id = $1 AND p.owner_user_id = $2 AND p.deleted_at IS NULL
		ORDER BY d.id DESC
	`, passengerID, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("list passenger documents: %w", err)
	}
	defer rows.Close()
	return collect(rows)
}

func collect(rows *sql.Rows) ([]documents.Document, error) {
	result := []documents.Document{}
	for rows.Next() {
		document, err := scanDocument(rows)
		if err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		result = append(result, *document)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate documents: %w", err)
	}
	return result, nil
}
