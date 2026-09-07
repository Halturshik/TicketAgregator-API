package service

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/bookingaccess"
	"github.com/google/uuid"
)

type Service struct {
	repo           bookingaccess.Repository
	challenges     bookingaccess.ChallengeStore
	access         bookingaccess.AccessStore
	mailer         bookingaccess.Mailer
	codes          bookingaccess.CodeGenerator
	refunds        bookingaccess.RefundService
	now            func() time.Time
	newChallengeID func() string
}

var _ bookingaccess.Service = (*Service)(nil)

func NewService(
	repo bookingaccess.Repository,
	challenges bookingaccess.ChallengeStore,
	access bookingaccess.AccessStore,
	mailer bookingaccess.Mailer,
	codes bookingaccess.CodeGenerator,
	refundService bookingaccess.RefundService,
) bookingaccess.Service {
	return &Service{
		repo:           repo,
		challenges:     challenges,
		access:         access,
		mailer:         mailer,
		codes:          codes,
		refunds:        refundService,
		now:            time.Now,
		newChallengeID: uuid.NewString,
	}
}
