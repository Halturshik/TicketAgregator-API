package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
)

type Repository struct {
	DB *sql.DB
}

var _ auth.UserStore = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}
