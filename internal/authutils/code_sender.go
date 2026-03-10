package authutils

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/interfaces"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
)

type CodeSender struct {
	codeStore interfaces.CodeStore
	mailer    interfaces.Mailer
}

func NewCodeSender(codeStore interfaces.CodeStore, mailer interfaces.Mailer) *CodeSender {
	return &CodeSender{
		codeStore: codeStore,
		mailer:    mailer,
	}
}

func (cs *CodeSender) Send(ctx context.Context, email string) error {
	code, err := cs.codeStore.Generate(ctx, email)
	if err != nil {
		logger.Error("Ошибка при генерации кода для %s: %v", email, err)
		return err
	}

	if err := cs.mailer.SendVerificationEmail(email, code); err != nil {
		logger.Error("Ошибка при отправке письма для %s: %v", email, err)
		return apierror.Wrap(err, apierror.ErrInternal)
	}

	logger.Info("Код подтверждения отправлен на %s", email)
	return nil
}
