package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/trips"
)

type Repository struct {
	db *sql.DB
}

var _ trips.Repository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
