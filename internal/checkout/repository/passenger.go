package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func (t *transaction) FindByDocument(ctx context.Context, ownerUserID int, fingerprint string) (int, error) {
	var passengerID int
	err := t.tx.QueryRowContext(ctx, `
		SELECT p.id
		FROM saved_passengers p
		JOIN documents d ON d.passenger_id = p.id
		WHERE p.owner_user_id = $1 AND p.deleted_at IS NULL AND d.document_fingerprint = $2
		LIMIT 1
		FOR UPDATE OF p
	`, ownerUserID, fingerprint).Scan(&passengerID)
	if err == sql.ErrNoRows {
		return 0, passengers.ErrNotFound
	}
	return passengerID, err
}

func (t *transaction) InsertFromPayment(ctx context.Context, ownerUserID int, passenger passengers.PaymentSnapshot) (int, error) {
	birthDate, err := time.Parse(documents.DateLayout, passenger.BirthDate)
	if err != nil {
		return 0, fmt.Errorf("parse passenger birth date: %w", err)
	}
	var passengerID int
	err = t.tx.QueryRowContext(ctx, `
		INSERT INTO saved_passengers
			(owner_user_id, first_name, middle_name, last_name, birth_date, is_russian)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6)
		RETURNING id
	`, ownerUserID, passenger.FirstName, passenger.MiddleName, passenger.LastName, birthDate, passenger.IsRussian).Scan(&passengerID)
	if err != nil {
		return 0, fmt.Errorf("insert saved passenger: %w", err)
	}
	return passengerID, nil
}

func (t *transaction) UpdateFromPayment(ctx context.Context, ownerUserID int, passengerID int, passenger passengers.PaymentSnapshot) error {
	birthDate, err := time.Parse(documents.DateLayout, passenger.BirthDate)
	if err != nil {
		return fmt.Errorf("parse passenger birth date: %w", err)
	}
	result, err := t.tx.ExecContext(ctx, `
		UPDATE saved_passengers
		SET first_name = $1, middle_name = NULLIF($2, ''), last_name = $3,
			birth_date = $4, is_russian = $5, updated_at = NOW()
		WHERE id = $6 AND owner_user_id = $7 AND deleted_at IS NULL
	`, passenger.FirstName, passenger.MiddleName, passenger.LastName, birthDate, passenger.IsRussian, passengerID, ownerUserID)
	if err != nil {
		return fmt.Errorf("update saved passenger: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected != 1 {
		return passengers.ErrNotFound
	}
	return nil
}
