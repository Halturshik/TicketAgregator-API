package repository

import (
	"context"
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func (r *Repository) GetOwned(ctx context.Context, ownerUserID int, documentID int, passengerID *int) (*documents.Document, error) {
	var nullablePassengerID sql.NullInt64
	if passengerID != nil {
		nullablePassengerID = sql.NullInt64{Int64: int64(*passengerID), Valid: true}
	}
	document, err := scanDocument(r.DB.QueryRowContext(ctx, `
		SELECT d.id, d.owner_user_id, d.passenger_id, d.type, d.number, d.verification_status,
			d.document_fingerprint, d.expires_at, d.last_checked_at
		FROM documents d
		LEFT JOIN saved_passengers p ON p.id = d.passenger_id
		WHERE d.id = $1 AND (
			($3::INT IS NULL AND d.owner_user_id = $2)
			OR ($3::INT IS NOT NULL AND d.passenger_id = $3 AND p.owner_user_id = $2 AND p.deleted_at IS NULL)
		)
	`, documentID, ownerUserID, nullablePassengerID))
	if err == sql.ErrNoRows {
		return nil, documents.ErrNotFound
	}
	return document, err
}
