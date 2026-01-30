package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/database/model"
	"github.com/Halturshik/TicketAgregator-API/internal/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/validator"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidFirstName  = errors.New("Некорректно указано имя")
	ErrInvalidMiddleName = errors.New("Некорректно указано отчество")
	ErrInvalidLastName   = errors.New("Некорректно указана фамилия")
	ErrInvalidEmail      = errors.New("Некорректный адрес электронной почты")
	ErrInvalidBirthDate  = errors.New("Некорректная дата рождения. Регистрация доступна для лиц, достигших 14 лет")
	ErrCodeExpired       = errors.New("Срок действия кода истек")
	ErrInvalidCode       = errors.New("Неверный код подтверждения")
	ErrInvalidPassword   = errors.New("Пароль должен состоять из 8 и более символов и содержать как минимум одну латинскую букву и одну цифру")
	ErrEmailIsUsed       = errors.New("Пользователь с таким email уже существует")
)

const emailVerify = "email_verify:"

type UserStore interface {
	IsEmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, u model.CreateUserParams) (int64, error)
}

type Mailer interface {
	SendVerificationEmail(to string, code string) error
}

type Service struct {
	store  UserStore
	mailer Mailer
	redis  *redis.Client
}

func NewService(store UserStore, mailer Mailer, redisClient *redis.Client) *Service {
	return &Service{
		store:  store,
		mailer: mailer,
		redis:  redisClient,
	}
}

type RegisterInput struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName   string `json:"last_name"`
	BirthDate  string `json:"birth_date"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsRussian  bool   `json:"is_russian"`
}

func (s *Service) StartRegistration(ctx context.Context, in RegisterInput) error {
	in.FirstName = cleaning.Name(in.FirstName)
	in.MiddleName = cleaning.Name(in.MiddleName)
	in.LastName = cleaning.Name(in.LastName)
	in.Email = cleaning.Email(in.Email)
	in.Password = cleaning.Password(in.Password)

	if !validator.NotEmpty(in.FirstName) || !validator.ValidName(in.FirstName) {
		return ErrInvalidFirstName
	}

	if validator.NotEmpty(in.MiddleName) && !validator.ValidName(in.MiddleName) {
		return ErrInvalidMiddleName
	}

	if !validator.NotEmpty(in.LastName) || !validator.ValidName(in.LastName) {
		return ErrInvalidLastName
	}

	if !validator.ValidEmail(in.Email) {
		return ErrInvalidEmail
	}

	if _, err := validator.ValidBirthDate(in.BirthDate); err != nil {
		return ErrInvalidBirthDate
	}

	if !validator.ValidPassword(in.Password) {
		return ErrInvalidPassword
	}

	exists, err := s.store.IsEmailExists(ctx, in.Email)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailIsUsed
	}

	code, err := GenerateVerificationCode()
	if err != nil {
		return err
	}

	err = s.redis.Set(ctx, emailVerify+in.Email, code, 2*time.Minute).Err()
	if err != nil {
		return err
	}

	return s.mailer.SendVerificationEmail(in.Email, code)
}

// нужно обсудить с Ксюшей, сможет ли она хранить и повторно прислывать данные с кодом, а то они сейчас теряются
func (s *Service) ConfirmRegistration(ctx context.Context, in RegisterInput, code string) error {
	storedCode, err := s.redis.Get(ctx, emailVerify+in.Email).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCodeExpired
		}
		return err
	}

	if storedCode != code {
		return ErrInvalidCode
	}

	birthDate, _ := validator.ValidBirthDate(in.BirthDate)

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	dbParams := model.CreateUserParams{
		FirstName:    in.FirstName,
		MiddleName:   in.MiddleName,
		LastName:     in.LastName,
		BirthDate:    birthDate,
		Email:        in.Email,
		PasswordHash: string(hash),
		IsRussian:    in.IsRussian,
	}

	_, err = s.store.CreateUser(ctx, dbParams)
	if err != nil {
		if err == errs.ErrDuplicateEmail {
			return ErrEmailIsUsed
		}
		return err
	}

	s.redis.Del(ctx, emailVerify+in.Email)

	return nil
}
