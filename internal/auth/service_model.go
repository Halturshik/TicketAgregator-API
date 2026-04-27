package auth

import (
	"github.com/Halturshik/TicketAgregator-API/internal/authutils"
	"github.com/Halturshik/TicketAgregator-API/internal/interfaces"
	"github.com/Halturshik/TicketAgregator-API/internal/redis"

	redisClient "github.com/redis/go-redis/v9"
)

type Service struct {
	store              interfaces.UserStore
	codeStore          interfaces.CodeStore
	codeSender         *authutils.CodeSender
	registrationStore  interfaces.RegistrationStore
	loginStore         interfaces.LoginStore
	refreshStore       interfaces.RefreshStore
	resetPasswordStore interfaces.ResetPasswordStore
}

func NewService(store interfaces.UserStore, mailer interfaces.Mailer, client *redisClient.Client) *Service {
	codeStore := redis.NewCodeService(client)
	codeSender := authutils.NewCodeSender(codeStore, mailer)
	registrationStore := redis.NewRegistrationStore(client)
	loginStore := redis.NewLoginStore(client)
	refreshStore := redis.NewRefreshStore(client)
	resetPasswordStore := redis.NewResetStore(client)
	return &Service{
		store:              store,
		codeSender:         codeSender,
		codeStore:          codeStore,
		registrationStore:  registrationStore,
		loginStore:         loginStore,
		refreshStore:       refreshStore,
		resetPasswordStore: resetPasswordStore,
	}
}
