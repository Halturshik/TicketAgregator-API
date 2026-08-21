package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

type scanner interface {
	Scan(dest ...any) error
}

func (r *Repository) FindRule(ctx context.Context, transport string, isInternational bool, age int, isRussian bool) (*documents.DocumentRule, error) {
	var rule documents.DocumentRule
	err := r.DB.QueryRowContext(ctx, `
		SELECT transport_type, is_international, min_age, max_age, is_russian,
			allow_internal_passport, allow_international_passport,
			allow_birth_certificate, allow_foreign_passport
		FROM document_rules
		WHERE transport_type IN ('any', $1)
			AND is_international = $2
			AND is_russian = $3
			AND min_age <= $4 AND max_age >= $4
		ORDER BY CASE WHEN transport_type = $1 THEN 0 ELSE 1 END
		LIMIT 1
	`, transport, isInternational, isRussian, age).Scan(
		&rule.Transport, &rule.IsInternational, &rule.MinAge, &rule.MaxAge, &rule.IsRussian,
		&rule.AllowInternalPassport, &rule.AllowInternationalPassport,
		&rule.AllowBirthCertificate, &rule.AllowForeignPassport,
	)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func scanDocument(s scanner) (*documents.Document, error) {
	var d documents.Document
	var ownerID sql.NullInt64
	var passengerID sql.NullInt64
	var expiresAt sql.NullTime
	if err := s.Scan(&d.ID, &ownerID, &passengerID, &d.Type, &d.Number, &d.VerificationStatus, &expiresAt); err != nil {
		return nil, err
	}
	if ownerID.Valid {
		v := int(ownerID.Int64)
		d.OwnerUserID = &v
	}
	if passengerID.Valid {
		v := int(passengerID.Int64)
		d.PassengerID = &v
	}
	if expiresAt.Valid {
		d.ExpiresAt = expiresAt.Time.Format("2006-01-02")
	}
	return &d, nil
}

func (r *Repository) ListForUser(ctx context.Context, ownerUserID int) ([]documents.Document, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, owner_user_id, passenger_id, type, number, verification_status, expires_at
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
		SELECT d.id, d.owner_user_id, d.passenger_id, d.type, d.number, d.verification_status, d.expires_at
		FROM documents d
		JOIN saved_passengers p ON p.id = d.passenger_id
		WHERE d.passenger_id = $1 AND p.owner_user_id = $2
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
		d, err := scanDocument(rows)
		if err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		result = append(result, *d)
	}
	return result, rows.Err()
}

func (r *Repository) CreateForUser(ctx context.Context, ownerUserID int, in documents.SaveDocumentInput, status string, expiresAt *time.Time) (*documents.Document, error) {
	return scanDocument(r.DB.QueryRowContext(ctx, `
		INSERT INTO documents (owner_user_id, type, number, verification_status, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, owner_user_id, passenger_id, type, number, verification_status, expires_at
	`, ownerUserID, in.Type, in.Number, status, expiresAt))
}

func (r *Repository) CreateForPassenger(ctx context.Context, ownerUserID int, passengerID int, in documents.SaveDocumentInput, status string, expiresAt *time.Time) (*documents.Document, error) {
	return scanDocument(r.DB.QueryRowContext(ctx, `
		INSERT INTO documents (passenger_id, type, number, verification_status, expires_at)
		SELECT p.id, $3, $4, $5, $6
		FROM saved_passengers p
		WHERE p.id = $1 AND p.owner_user_id = $2
		RETURNING id, owner_user_id, passenger_id, type, number, verification_status, expires_at
	`, passengerID, ownerUserID, in.Type, in.Number, status, expiresAt))
}
