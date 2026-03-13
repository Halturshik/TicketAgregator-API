package auth

import (
	"context"

	"github.com/Halturshik/TicketAgregator-API/database/errs"
	"github.com/Halturshik/TicketAgregator-API/database/model"
	"github.com/Halturshik/TicketAgregator-API/internal/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/auth/types"
	"github.com/Halturshik/TicketAgregator-API/internal/authutils"
	"github.com/Halturshik/TicketAgregator-API/internal/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/interfaces"
	"github.com/Halturshik/TicketAgregator-API/internal/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/redis"
	"github.com/Halturshik/TicketAgregator-API/internal/validator"

	redisClient "github.com/redis/go-redis/v9"
)

type Service struct {
	store             interfaces.UserStore
	codeStore         interfaces.CodeStore
	codeSender        *authutils.CodeSender
	registrationStore interfaces.RegistrationStore
	loginStore        interfaces.LoginStore
	refreshStore      interfaces.RefreshStore
}

func NewService(store interfaces.UserStore, mailer interfaces.Mailer, client *redisClient.Client) *Service {
	codeStore := redis.NewCodeService(client)
	codeSender := authutils.NewCodeSender(codeStore, mailer)
	registrationStore := redis.NewRegistrationStore(client)
	loginStore := redis.NewLoginStore(client)
	refreshStore := redis.NewRefreshStore(client)
	return &Service{
		store:             store,
		codeSender:        codeSender,
		codeStore:         codeStore,
		registrationStore: registrationStore,
		loginStore:        loginStore,
		refreshStore:      refreshStore,
	}
}

func (s *Service) StartRegistration(ctx context.Context, in types.RegisterInput) error {
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

	err = s.registrationStore.Save(ctx, in.Email, in)
	if err != nil {
		return err
	}

	if err := s.codeSender.Send(ctx, in.Email); err != nil {
		return err
	}

	return nil
}

func (s *Service) ConfirmRegistration(ctx context.Context, in types.ConfirmRegisterInput) (*types.LoginOutput, error) {
	in.Email = cleaning.Email(in.Email)

	if err := s.codeStore.Verify(ctx, in.Email, in.Code); err != nil {
		return nil, err
	}

	stored, err := s.registrationStore.Get(ctx, in.Email)
	if err != nil {
		return nil, err
	}

	birthDate, _ := validator.ValidBirthDate(stored.BirthDate)

	hash, err := authutils.HashPassword(stored.Password)
	if err != nil {
		logger.Error("Ошибка при хэшировании пароля для %s: %v", in.Email, err)
		return nil, err
	}

	dbParams := model.CreateUserParams{
		FirstName:    stored.FirstName,
		MiddleName:   stored.MiddleName,
		LastName:     stored.LastName,
		BirthDate:    birthDate,
		Email:        stored.Email,
		PasswordHash: hash,
		IsRussian:    stored.IsRussian,
	}

	userID, err := s.store.CreateUser(ctx, dbParams)
	if err != nil {
		if err == errs.ErrDuplicateEmail {
			return nil, apierror.ErrEmailIsUsed
		}
		logger.Error("Ошибка при создании пользователя %s: %v", in.Email, err)
		return nil, err
	}

	accessToken, err := authutils.GenerateToken(int(userID), authutils.AccessTokenTTL)
	if err != nil {
		logger.Error("Ошибка генерации access токена: %v", err)
		return nil, err
	}

	refreshToken, err := authutils.GenerateToken(int(userID), authutils.RefreshTokenTTL)
	if err != nil {
		logger.Error("Ошибка генерации refresh токена: %v", err)
		return nil, err
	}

	if err := s.refreshStore.Save(ctx, userID, refreshToken); err != nil {
		return nil, err
	}

	if err := s.codeStore.Clear(ctx, in.Email); err != nil {
		logger.Warn("Ошибка при инвалидации кода после регистрации %s: %v", in.Email, err)
		return nil, err
	}

	if err := s.registrationStore.Delete(ctx, in.Email); err != nil {
		logger.Warn("Ошибка при очистке регистрационных данных из временного хранилища %s: %v", in.Email, err)
		return nil, err
	}

	return &types.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
