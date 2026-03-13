package interfaces

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/database/model"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/types"
)

type UserStore interface {
	IsEmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, u model.CreateUserParams) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

type Mailer interface {
	SendVerificationEmail(to string, code string) error
}

type RegistrationStore interface {
	Save(ctx context.Context, email string, data types.RegisterInput) error
	Get(ctx context.Context, email string) (*types.RegisterInput, error)
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
}
