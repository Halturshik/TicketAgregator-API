package auth

import (
	"context"
	"time"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/database/model"
	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/validator"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const emailVerify = "email_verify:"

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

	fields := map[string]string{}

	if !validator.NotEmpty(in.FirstName) || !validator.ValidName(in.FirstName) {
		fields[apierror.FieldFirstName] = apierror.ErrInvalidFirstName
	}

	if validator.NotEmpty(in.MiddleName) && !validator.ValidName(in.MiddleName) {
		fields[apierror.FieldMiddleName] = apierror.ErrInvalidMiddleName
	}

	if !validator.NotEmpty(in.LastName) || !validator.ValidName(in.LastName) {
		fields[apierror.FieldLastName] = apierror.ErrInvalidLastName
	}

	if !validator.ValidEmail(in.Email) {
		fields[apierror.FieldEmail] = apierror.ErrInvalidEmail
	}

	if _, err := validator.ValidBirthDate(in.BirthDate); err != nil {
		fields[apierror.FieldBirthDate] = apierror.ErrInvalidBirthDate
	}

	if !validator.ValidPassword(in.Password) {
		fields[apierror.FieldPassword] = apierror.ErrInvalidPassword
	}

	if len(fields) > 0 {
		return apierror.Validation(fields)
	}

	exists, err := s.store.IsEmailExists(ctx, in.Email)
	if err != nil {
		return err
	}
	if exists {
		return apierror.ErrEmailIsUsed
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

// нужно обсудить с Ксюшей, сможет ли она хранить и повторно присылать данные с кодом, а то они сейчас теряются
func (s *Service) ConfirmRegistration(ctx context.Context, in RegisterInput, code string) error {
	storedCode, err := s.redis.Get(ctx, emailVerify+in.Email).Result()
	if err != nil {
		if err == redis.Nil {
			return apierror.ErrCodeExpired
		}
		return err
	}

	if storedCode != code {
		return apierror.ErrInvalidVerificationCode
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
			return apierror.ErrEmailIsUsed
		}
		return err
	}

	s.redis.Del(ctx, emailVerify+in.Email)

	return nil
}
