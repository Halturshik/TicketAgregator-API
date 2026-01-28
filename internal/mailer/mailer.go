package mailer

import (
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
)

type Mailer interface {
	SendVerificationEmail(to string, code string) error
}

type ConsoleMailer struct{}

func (c *ConsoleMailer) SendVerificationEmail(to string, code string) error {
	logger.Info("[FAKE EMAIL] To: %s | Verification code: %s\n", to, code)
	return nil
}
