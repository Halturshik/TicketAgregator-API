package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

func requireSingleRow(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return documents.ErrNotFound
	}
	return nil
}
