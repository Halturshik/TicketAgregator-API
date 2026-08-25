package repository

import (
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func documentDates(document documents.PaymentDocument) (any, time.Time, error) {
	expiresAt, err := nullableDate(document.ExpiresAt)
	if err != nil {
		return nil, time.Time{}, err
	}
	checkedAt, err := time.Parse(time.RFC3339, document.LastCheckedAt)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("parse document last checked time: %w", err)
	}
	return expiresAt, checkedAt, nil
}

func nullableDate(value string) (any, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(documents.DateLayout, value)
	if err != nil {
		return nil, fmt.Errorf("parse document expiration date: %w", err)
	}
	return parsed, nil
}
