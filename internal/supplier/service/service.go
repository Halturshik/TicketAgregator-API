package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
	"github.com/google/uuid"
)

type Service struct {
	repo  supplier.RefundRepository
	now   func() time.Time
	newID func() string
}

var _ supplier.Service = (*Service)(nil)

func NewService(repo supplier.RefundRepository) supplier.Service {
	return &Service{repo: repo, now: time.Now, newID: uuid.NewString}
}
