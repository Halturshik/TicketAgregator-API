package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/cleaning"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
	"github.com/Halturshik/TicketAgregator-API/internal/passengers/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) passengers.Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, ownerUserID int) ([]passengers.Passenger, error) {
	items, err := s.repo.List(ctx, ownerUserID)
	if err != nil {
		logger.Error("Ошибка получения сохраненных пассажиров userID=%d: %v", ownerUserID, err)
		return nil, err
	}
	return items, nil
}

func (s *Service) Create(ctx context.Context, ownerUserID int, in passengers.SavePassengerInput) (*passengers.Passenger, error) {
	p, err := s.validate(&in)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.Create(ctx, ownerUserID, in, p)
	if err != nil {
		logger.Error("Ошибка создания сохраненного пассажира userID=%d: %v", ownerUserID, err)
		return nil, err
	}
	logger.Info("Сохраненный пассажир создан: userID=%d passengerID=%d", ownerUserID, item.ID)
	return item, nil
}

func (s *Service) Update(ctx context.Context, ownerUserID int, passengerID int, in passengers.SavePassengerInput) (*passengers.Passenger, error) {
	p, err := s.validate(&in)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.Update(ctx, ownerUserID, passengerID, in, p)
	if err == sql.ErrNoRows {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка обновления сохраненного пассажира userID=%d passengerID=%d: %v", ownerUserID, passengerID, err)
		return nil, err
	}
	logger.Info("Сохраненный пассажир обновлен: userID=%d passengerID=%d", ownerUserID, passengerID)
	return item, nil
}

func (s *Service) validate(in *passengers.SavePassengerInput) (time.Time, error) {
	in.FirstName = cleaning.Name(in.FirstName)
	in.MiddleName = cleaning.Name(in.MiddleName)
	in.LastName = cleaning.Name(in.LastName)

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
	parsedBirthDate, parseErr := validPassengerBirthDate(in.BirthDate)
	if parseErr != nil {
		fields[apierror.FieldBirthDate] = apierror.ErrInvalidBirthDate
	}
	if len(fields) > 0 {
		return time.Time{}, apierror.Validation(fields)
	}
	return parsedBirthDate, nil
}

func validPassengerBirthDate(value string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, err
	}
	if t.After(time.Now()) {
		return time.Time{}, apierror.ErrInvalidRequest
	}
	return t, nil
}
