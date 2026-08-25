package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/users"
)

type Repository struct {
	DB *sql.DB
}

var _ users.Repository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}
