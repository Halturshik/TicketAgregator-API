package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

type Repository struct {
	DB *sql.DB
}

var _ search.Repository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}
