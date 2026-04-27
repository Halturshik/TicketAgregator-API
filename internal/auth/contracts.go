package auth

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/repository"
)

type UserStore interface {
	IsEmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, u repository.CreateUserParams) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*repository.UserAuth, error)
	GetUserByID(ctx context.Context, id int) (*repository.UserAuth, error)
	UpdatePassword(ctx context.Context, userID int, hash string) (int, error)
}

type Mailer interface {
	SendVerificationEmail(to string, code string) error
}

type RegistrationStore interface {
	Save(ctx context.Context, email string, data RegisterInput) error
	Get(ctx context.Context, email string) (*RegisterInput, error)
	Delete(ctx context.Context, email string) error
}

type LoginStore interface {
	Save(ctx context.Context, email string, userID int64) error
	Get(ctx context.Context, email string) (int64, error)
	Delete(ctx context.Context, email string) error
}

type CodeStore interface {
	Generate(ctx context.Context, email string) (string, error)
	Verify(ctx context.Context, email, code string) error
	Clear(ctx context.Context, email string) error
}

type RefreshStore interface {
	Save(ctx context.Context, userID int64, token string) error
	Get(ctx context.Context, token string) (int64, error)
	Delete(ctx context.Context, token string) error
	AddToUserSet(ctx context.Context, userID int64, tokenHash string) error
	RemoveFromUserSet(ctx context.Context, userID int64, tokenHash string) error
	DeleteAllForUser(ctx context.Context, userID int64) error
}

type ResetPasswordStore interface {
	SaveVerified(ctx context.Context, email string) error
	IsVerified(ctx context.Context, email string) (bool, error)
	Delete(ctx context.Context, email string) error
}
