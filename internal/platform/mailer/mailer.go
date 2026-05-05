package mailer

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

type ConsoleMailer struct{}

func (c *ConsoleMailer) SendVerificationEmail(ctx context.Context, to string, code string) error {
	logger.Info("[FAKE EMAIL] To: %s | Verification code: %s\n", to, code)
	return nil
}
