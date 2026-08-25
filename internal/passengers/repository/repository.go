package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

type Repository struct {
	DB *sql.DB
}

var _ passengers.Repository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}
