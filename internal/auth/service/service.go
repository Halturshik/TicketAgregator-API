package service

import (
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/code"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/store"
	redisClient "github.com/redis/go-redis/v9"
)

type Service struct {
	store              auth.UserStore
	codeStore          auth.CodeStore
	codeSender         *code.CodeSender
	registrationStore  auth.RegistrationStore
	loginStore         auth.LoginStore
	refreshStore       auth.RefreshStore
	resetPasswordStore auth.ResetPasswordStore
}

func NewService(userStore auth.UserStore, mailer auth.Mailer, client *redisClient.Client) auth.AuthService {
	codeStore := store.NewCodeService(client)
	codeSender := code.NewCodeSender(codeStore, mailer)
	registrationStore := store.NewRegistrationStore(client)
	loginStore := store.NewLoginStore(client)
	refreshStore := store.NewRefreshStore(client)
	resetPasswordStore := store.NewResetStore(client)
	return &Service{
		store:              userStore,
		codeSender:         codeSender,
		codeStore:          codeStore,
		registrationStore:  registrationStore,
		loginStore:         loginStore,
		refreshStore:       refreshStore,
		resetPasswordStore: resetPasswordStore,
	}
}
