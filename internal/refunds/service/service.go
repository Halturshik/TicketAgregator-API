package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/refunds"
	"github.com/Halturshik/TicketAgregator-API/internal/supplier"
)

type Service struct {
	repo     refunds.Repository
	supplier supplier.Gateway
	now      func() time.Time
}

var _ refunds.Service = (*Service)(nil)

func NewService(repo refunds.Repository, supplierGateway supplier.Gateway) refunds.Service {
	return &Service{repo: repo, supplier: supplierGateway, now: time.Now}
}
