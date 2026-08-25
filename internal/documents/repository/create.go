package repository

import (
	"context"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func (r *Repository) CreateForUser(
	ctx context.Context,
	ownerUserID int,
	in documents.SaveDocumentInput,
	status string,
	fingerprint string,
	expiresAt *time.Time,
	checkedAt time.Time,
) (*documents.Document, error) {
	document, err := scanDocument(r.DB.QueryRowContext(ctx, `
		INSERT INTO documents
			(owner_user_id, type, number, verification_status, document_fingerprint, expires_at, last_checked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, owner_user_id, passenger_id, type, number, verification_status,
			document_fingerprint, expires_at, last_checked_at
	`, ownerUserID, in.Type, in.Number, status, fingerprint, expiresAt, checkedAt))
	if err != nil {
		return nil, mapDocumentWriteError(err)
	}
	return document, nil
}

func (r *Repository) CreateForPassenger(
	ctx context.Context,
	ownerUserID int,
	passengerID int,
	in documents.SaveDocumentInput,
	status string,
	fingerprint string,
	expiresAt *time.Time,
	checkedAt time.Time,
) (*documents.Document, error) {
	document, err := scanDocument(r.DB.QueryRowContext(ctx, `
		INSERT INTO documents
			(passenger_id, type, number, verification_status, document_fingerprint, expires_at, last_checked_at)
		SELECT p.id, $3, $4, $5, $6, $7, $8
		FROM saved_passengers p
		WHERE p.id = $1 AND p.owner_user_id = $2 AND p.deleted_at IS NULL
		RETURNING id, owner_user_id, passenger_id, type, number, verification_status,
			document_fingerprint, expires_at, last_checked_at
	`, passengerID, ownerUserID, in.Type, in.Number, status, fingerprint, expiresAt, checkedAt))
	if err != nil {
		return nil, mapDocumentWriteError(err)
	}
	return document, nil
}
