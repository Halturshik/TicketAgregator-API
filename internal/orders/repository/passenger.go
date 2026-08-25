package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

func insertOrderPassengers(ctx context.Context, tx *sql.Tx, orderID int, drafts []orders.OrderPassengerDraft) ([]int, error) {
	ids := make([]int, 0, len(drafts))
	for _, draft := range drafts {
		id, err := insertOrderPassenger(ctx, tx, orderID, draft)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func insertOrderPassenger(ctx context.Context, tx *sql.Tx, orderID int, draft orders.OrderPassengerDraft) (int, error) {
	passengerData, err := json.Marshal(draft.Passenger)
	if err != nil {
		return 0, fmt.Errorf("marshal passenger snapshot: %w", err)
	}
	documentData, err := json.Marshal(draft.Document)
	if err != nil {
		return 0, fmt.Errorf("marshal document snapshot: %w", err)
	}

	var savedPassengerID sql.NullInt64
	if draft.SavedPassengerID != nil {
		savedPassengerID = sql.NullInt64{Int64: int64(*draft.SavedPassengerID), Valid: true}
	}
	var savedDocumentID sql.NullInt64
	if draft.SavedDocumentID != nil {
		savedDocumentID = sql.NullInt64{Int64: int64(*draft.SavedDocumentID), Valid: true}
	}

	var id int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO order_passengers
			(order_id, passenger_source, saved_passenger_id, saved_document_id, save_changes, passenger_snapshot, document_snapshot, document_fingerprint)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, orderID, draft.Source, savedPassengerID, savedDocumentID, draft.SaveChanges, passengerData, documentData, draft.DocumentFingerprint).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert order passenger: %w", err)
	}
	return id, nil
}
