package auth

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/internal/auth/repository"
)

type AuthService interface {
	StartRegistration(ctx context.Context, in RegisterInput) error
	ConfirmRegistration(ctx context.Context, in ConfirmRegisterInput) (*LoginOutput, error)

	LoginStart(ctx context.Context, in LoginStartInput) error
	LoginConfirm(ctx context.Context, in LoginConfirmInput) (*LoginOutput, error)

	Refresh(ctx context.Context, refreshToken string) (*LoginOutput, error)
	Logout(ctx context.Context, refreshToken string) error

	ForgotPassword(ctx context.Context, email string) error
	VerifyResetCode(ctx context.Context, in PasswordVerifyInput) error
	ResetPassword(ctx context.Context, in ChangePasswordConfirmInput) (*LoginOutput, error)
}

type UserStore interface {
	IsEmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, u repository.CreateUserParams) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*repository.UserAuth, error)
	GetUserByID(ctx context.Context, id int) (*repository.UserAuth, error)
	UpdatePassword(ctx context.Context, userID int, hash string) (int, error)
}

type Mailer interface {
	SendVerificationEmail(ctx context.Context, to string, code string) error
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
	RequestCode(ctx context.Context, email string, code string) error
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

type TokenManager interface {
	GenerateAccessToken(userID int, version int) (string, error)
	GenerateRefreshToken(userID int, version int) (string, error)
	ParseToken(tokenStr string) (int, int, error)
}

type CodeGenerator interface {
	GenerateVerificationCode() (string, error)
}
