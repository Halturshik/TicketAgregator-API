package service

import (
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/code"
)

type Service struct {
	store              auth.UserStore
	codeStore          auth.CodeStore
	codeSender         *code.CodeSender
	registrationStore  auth.RegistrationStore
	loginStore         auth.LoginStore
	refreshStore       auth.RefreshStore
	resetPasswordStore auth.ResetPasswordStore
	jwt                auth.TokenManager
}

func NewService(
	userStore auth.UserStore,
	codeSender *code.CodeSender,
	codeStore auth.CodeStore,
	registrationStore auth.RegistrationStore,
	loginStore auth.LoginStore,
	refreshStore auth.RefreshStore,
	resetPasswordStore auth.ResetPasswordStore,
	jwt auth.TokenManager,
) auth.AuthService {

	return &Service{
		store:              userStore,
		codeSender:         codeSender,
		codeStore:          codeStore,
		registrationStore:  registrationStore,
		loginStore:         loginStore,
		refreshStore:       refreshStore,
		resetPasswordStore: resetPasswordStore,
		jwt:                jwt,
	}
}
