package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

type Repository struct {
	db *sql.DB
}

var _ supplier.RefundRepository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
