package auth

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/database/model"
	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/validator"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	store       UserStore
	mailer      Mailer
	codeService *CodeService
}

func NewService(store UserStore, mailer Mailer, redisClient *redis.Client) *Service {
	return &Service{
		store:       store,
		mailer:      mailer,
		codeService: NewCodeService(redisClient),
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
		logger.Error("Ошибка при проверке существования email %s: %v", in.Email, err)
		return err
	}
	if exists {
		logger.Warn("Попытка регистрации на уже существующий email: %s", in.Email)
		return apierror.ErrEmailIsUsed
	}

	code, err := s.codeService.Generate(ctx, in.Email)
	if err != nil {
		logger.Warn("Ошибка при генерации кода для %s: %v", in.Email, err)
		return err
	}

	if err := s.mailer.SendVerificationEmail(in.Email, code); err != nil {
		logger.Warn("Ошибка при отправке письма для %s: %v", in.Email, err)
		return err
	}

	logger.Info("Код подтверждения отправлен на %s", in.Email)
	return nil
}

// нужно обсудить с Ксюшей, сможет ли она хранить и повторно присылать данные с кодом, а то они сейчас теряются
func (s *Service) ConfirmRegistration(ctx context.Context, in RegisterInput, code string) error {
	in.Email = cleaning.Email(in.Email)

	if err := s.codeService.Verify(ctx, in.Email, code); err != nil {
		return err
	}

	birthDate, _ := validator.ValidBirthDate(in.BirthDate)

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Ошибка при хэшировании пароля для %s: %v", in.Email, err)
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
		logger.Error("Ошибка при создании пользователя %s: %v", in.Email, err)
		return err
	}

	if err := s.codeService.Clear(ctx, in.Email); err != nil {
		logger.Warn("Ошибка при инвалидации кода после регистрации %s: %v", in.Email, err)
		return err
	}

	logger.Info("Успешная регистрация пользователя: %s", in.Email)
	return nil
}
