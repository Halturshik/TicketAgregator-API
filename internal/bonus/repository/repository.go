package repository

import (
	"database/sql"

	"github.com/Halturshik/TicketAgregator-API/internal/bonus"
)

type Repository struct {
	DB *sql.DB
}

var _ bonus.Repository = (*Repository)(nil)
var _ bonus.BalanceReader = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}
