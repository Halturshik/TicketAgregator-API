package auth

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail     = errors.New("некорректный адрес электронной почты")
	ErrInvalidBirthDate = errors.New("некорректная дата рождения") // заглушка, потом поправить
)

type UserStore interface {
	CreateUser(ctx context.Context, u CreateUserParams) (int64, error)
}

type Service struct {
	store UserStore
}

func NewService(store UserStore) *Service {
	return &Service{store: store}
}

type RegisterInput struct {
	FirstName  string
	MiddleName string
	LastName   string
	BirthDate  string
	Email      string
	Password   string
	IsRussian  bool
}

type CreateUserParams struct {
	FirstName    string
	MiddleName   string
	LastName     string
	BirthDate    time.Time
	Email        string
	PasswordHash string
	IsRussian    bool
}

func (s *Service) Register(ctx context.Context, in RegisterInput) error {
	if !validator.ValidEmail(in.Email) {
		return ErrInvalidEmail
	}

	birthDate, err := time.Parse("2006-01-02", in.BirthDate)
	if err != nil {
		return ErrInvalidBirthDate
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.store.CreateUser(ctx, CreateUserParams{
		FirstName:    in.FirstName,
		MiddleName:   in.MiddleName,
		LastName:     in.LastName,
		BirthDate:    birthDate,
		Email:        in.Email,
		PasswordHash: string(hash),
		IsRussian:    in.IsRussian,
	})

	return err
}
