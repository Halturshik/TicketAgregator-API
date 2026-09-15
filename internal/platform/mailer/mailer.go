package mailer

import (
	"context"
	"log/slog"

	platformlogger "github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

type ConsoleMailer struct{}

func (c *ConsoleMailer) SendVerificationEmail(ctx context.Context, to string, code string) error {
	slog.DebugContext(ctx, "Отправлено mock-письмо с кодом подтверждения",
		slog.String("email", platformlogger.MaskEmail(to)),
		slog.String("verification_code", code),
		slog.Bool("mock", true),
	)
	return nil
}
