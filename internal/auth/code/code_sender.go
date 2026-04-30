package code

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

type CodeSender struct {
	codeStore auth.CodeStore
	mailer    auth.Mailer
}

func NewCodeSender(codeStore auth.CodeStore, mailer auth.Mailer) *CodeSender {
	return &CodeSender{
		codeStore: codeStore,
		mailer:    mailer,
	}
}

func (cs *CodeSender) Send(ctx context.Context, email string) error {
	code, err := cs.codeStore.Generate(ctx, email)
	if err != nil {
		logger.Warn("Ошибка при генерации кода для %s: %v", email, err)
		return err
	}

	if err := cs.mailer.SendVerificationEmail(email, code); err != nil {
		logger.Error("Ошибка при отправке письма для %s: %v", email, err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	return nil
}
