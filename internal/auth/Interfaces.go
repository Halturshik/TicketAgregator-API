package auth

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/database/model"
)

type UserStore interface {
	IsEmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, u model.CreateUserParams) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

type Mailer interface {
	SendVerificationEmail(to string, code string) error
}
