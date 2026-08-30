package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
)

type Repository struct {
	db *sql.DB
}

var _ refunds.Repository = (*Repository)(nil)
var _ refunds.Transaction = (*transaction)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
