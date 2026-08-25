package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

type Repository struct {
	DB *sql.DB
}

var _ documents.Repository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}
