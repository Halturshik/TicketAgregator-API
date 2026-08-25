package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/checkout"
)

type Repository struct {
	DB *sql.DB
}

var _ checkout.Repository = (*Repository)(nil)
var _ checkout.Transaction = (*transaction)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}
