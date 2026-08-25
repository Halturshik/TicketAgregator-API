package service

import "github.com/Halturshik/TicketAgregator-API/internal/auth"

type Service struct {
	store              auth.UserStore
	codeStore          auth.CodeStore
	mailer             auth.Mailer
	codeGenerator      auth.CodeGenerator
	registrationStore  auth.RegistrationStore
	loginStore         auth.LoginStore
	refreshStore       auth.RefreshStore
	resetPasswordStore auth.ResetPasswordStore
	jwt                auth.TokenManager
}

func NewService(
	userStore auth.UserStore,
	mailer auth.Mailer,
	codeGenerator auth.CodeGenerator,
	codeStore auth.CodeStore,
	registrationStore auth.RegistrationStore,
	loginStore auth.LoginStore,
	refreshStore auth.RefreshStore,
	resetPasswordStore auth.ResetPasswordStore,
	jwt auth.TokenManager,
) auth.Service {
	return &Service{
		store:              userStore,
		mailer:             mailer,
		codeGenerator:      codeGenerator,
		codeStore:          codeStore,
		registrationStore:  registrationStore,
		loginStore:         loginStore,
		refreshStore:       refreshStore,
		resetPasswordStore: resetPasswordStore,
		jwt:                jwt,
	}
}
