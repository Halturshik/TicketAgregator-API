package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

type Service struct {
	repo               documents.Repository
	verificationSecret []byte
	now                func() time.Time
}

func NewService(repo documents.Repository, verificationSecret string) documents.Service {
	return &Service{
		repo:               repo,
		verificationSecret: []byte(verificationSecret),
		now:                time.Now,
	}
}
