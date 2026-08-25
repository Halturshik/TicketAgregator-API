package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

func (t *transaction) ListPassengersForPayment(ctx context.Context, orderID int) ([]passengers.PaymentPassenger, error) {
	rows, err := t.tx.QueryContext(ctx, `
		SELECT passenger_source, saved_passenger_id, saved_document_id, save_changes,
			passenger_snapshot, document_snapshot, document_fingerprint
		FROM order_passengers
		WHERE order_id = $1
		ORDER BY id
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("list order passengers: %w", err)
	}
	defer rows.Close()

	items := make([]passengers.PaymentPassenger, 0)
	for rows.Next() {
		item, err := scanPaymentPassenger(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order passengers: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close order passengers: %w", err)
	}
	return items, nil
}

func scanPaymentPassenger(rows *sql.Rows) (*passengers.PaymentPassenger, error) {
	var item passengers.PaymentPassenger
	var savedPassengerID sql.NullInt64
	var savedDocumentID sql.NullInt64
	var passengerData []byte
	var documentData []byte
	if err := rows.Scan(
		&item.Source, &savedPassengerID, &savedDocumentID, &item.SaveChanges,
		&passengerData, &documentData, &item.Document.Fingerprint,
	); err != nil {
		return nil, fmt.Errorf("scan order passenger: %w", err)
	}
	if err := json.Unmarshal(passengerData, &item.Passenger); err != nil {
		return nil, fmt.Errorf("unmarshal passenger snapshot: %w", err)
	}
	if err := json.Unmarshal(documentData, &item.Document); err != nil {
		return nil, fmt.Errorf("unmarshal document snapshot: %w", err)
	}
	if savedPassengerID.Valid {
		value := int(savedPassengerID.Int64)
		item.SavedPassengerID = &value
	}
	if savedDocumentID.Valid {
		value := int(savedDocumentID.Int64)
		item.Document.SavedDocumentID = &value
	}
	return &item, nil
}
