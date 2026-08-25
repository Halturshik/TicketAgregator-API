package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func (r *Repository) Update(
	ctx context.Context,
	ownerUserID int,
	documentID int,
	in documents.SaveDocumentInput,
	status string,
	fingerprint string,
	expiresAt *time.Time,
	checkedAt time.Time,
) (*documents.Document, error) {
	var nullablePassengerID sql.NullInt64
	if in.PassengerID != nil {
		nullablePassengerID = sql.NullInt64{Int64: int64(*in.PassengerID), Valid: true}
	}
	document, err := scanDocument(r.DB.QueryRowContext(ctx, `
		UPDATE documents d
		SET type = $4, number = $5, verification_status = $6, document_fingerprint = $7,
			expires_at = $8, last_checked_at = $9, updated_at = NOW()
		WHERE d.id = $1 AND (
			($3::INT IS NULL AND d.owner_user_id = $2)
			OR ($3::INT IS NOT NULL AND d.passenger_id = $3 AND EXISTS (
				SELECT 1 FROM saved_passengers p
				WHERE p.id = $3 AND p.owner_user_id = $2 AND p.deleted_at IS NULL
			))
		)
		RETURNING id, owner_user_id, passenger_id, type, number, verification_status,
			document_fingerprint, expires_at, last_checked_at
	`, documentID, ownerUserID, nullablePassengerID, in.Type, in.Number, status, fingerprint, expiresAt, checkedAt))
	if err != nil {
		return nil, mapDocumentWriteError(err)
	}
	return document, nil
}

func (r *Repository) UpdateVerification(ctx context.Context, documentID int, status string, checkedAt time.Time) error {
	result, err := r.DB.ExecContext(ctx, `
		UPDATE documents
		SET verification_status = $2, last_checked_at = $3, updated_at = NOW()
		WHERE id = $1
	`, documentID, status, checkedAt)
	if err != nil {
		return err
	}
	return requireSingleRow(result)
}
