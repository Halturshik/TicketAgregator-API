package repository

import (
	"database/sql"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanDocument(source scanner) (*documents.Document, error) {
	var document documents.Document
	var ownerID sql.NullInt64
	var passengerID sql.NullInt64
	var expiresAt sql.NullTime
	var lastCheckedAt time.Time
	if err := source.Scan(
		&document.ID, &ownerID, &passengerID, &document.Type, &document.Number, &document.VerificationStatus,
		&document.Fingerprint, &expiresAt, &lastCheckedAt,
	); err != nil {
		return nil, err
	}
	if ownerID.Valid {
		value := int(ownerID.Int64)
		document.OwnerUserID = &value
	}
	if passengerID.Valid {
		value := int(passengerID.Int64)
		document.PassengerID = &value
	}
	if expiresAt.Valid {
		document.ExpiresAt = expiresAt.Time.Format(documents.DateLayout)
	}
	document.LastCheckedAt = lastCheckedAt.Format(time.RFC3339)
	return &document, nil
}
