package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func (t *transaction) UpsertUserAfterPayment(ctx context.Context, ownerUserID int, document documents.PaymentDocument) error {
	expiresAt, checkedAt, err := documentDates(document)
	if err != nil {
		return err
	}
	_, err = t.tx.ExecContext(ctx, `
		INSERT INTO documents
			(owner_user_id, type, number, verification_status, document_fingerprint, expires_at, last_checked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (owner_user_id, document_fingerprint) WHERE owner_user_id IS NOT NULL
		DO UPDATE SET type = EXCLUDED.type, number = EXCLUDED.number,
			verification_status = EXCLUDED.verification_status, expires_at = EXCLUDED.expires_at,
			last_checked_at = EXCLUDED.last_checked_at, updated_at = NOW()
	`, ownerUserID, document.Type, document.Number, document.VerificationStatus,
		document.Fingerprint, expiresAt, checkedAt)
	if err != nil {
		return fmt.Errorf("upsert user document: %w", err)
	}
	return nil
}

func (t *transaction) UpdatePassengerAfterPayment(ctx context.Context, passengerID int, document documents.PaymentDocument) (bool, error) {
	if document.SavedDocumentID == nil {
		return false, fmt.Errorf("saved document id is required")
	}
	expiresAt, checkedAt, err := documentDates(document)
	if err != nil {
		return false, err
	}
	result, err := t.tx.ExecContext(ctx, `
		UPDATE documents
		SET type = $1, number = $2, verification_status = $3, document_fingerprint = $4,
			expires_at = $5, last_checked_at = $6, updated_at = NOW()
		WHERE id = $7 AND passenger_id = $8
	`, document.Type, document.Number, document.VerificationStatus,
		document.Fingerprint, expiresAt, checkedAt,
		*document.SavedDocumentID, passengerID)
	if err != nil {
		return false, fmt.Errorf("update saved passenger document: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (t *transaction) UpsertPassengerAfterPayment(ctx context.Context, passengerID int, document documents.PaymentDocument) error {
	expiresAt, checkedAt, err := documentDates(document)
	if err != nil {
		return err
	}
	_, err = t.tx.ExecContext(ctx, `
		INSERT INTO documents
			(passenger_id, type, number, verification_status, document_fingerprint, expires_at, last_checked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (passenger_id, document_fingerprint) WHERE passenger_id IS NOT NULL
		DO UPDATE SET type = EXCLUDED.type, number = EXCLUDED.number,
			verification_status = EXCLUDED.verification_status, expires_at = EXCLUDED.expires_at,
			last_checked_at = EXCLUDED.last_checked_at, updated_at = NOW()
	`, passengerID, document.Type, document.Number, document.VerificationStatus,
		document.Fingerprint, expiresAt, checkedAt)
	if err != nil {
		return fmt.Errorf("upsert saved passenger document: %w", err)
	}
	return nil
}
