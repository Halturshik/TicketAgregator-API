package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/orders"
)

type Repository struct {
	DB *sql.DB
}

var _ orders.Repository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}
