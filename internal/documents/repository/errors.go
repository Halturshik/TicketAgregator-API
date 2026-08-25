package repository

import (
	"database/sql"
	"errors"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/lib/pq"
)

const uniqueViolationCode = "23505"

func mapDocumentWriteError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return documents.ErrNotFound
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == uniqueViolationCode {
		return documents.ErrAlreadyExists
	}
	return err
}
