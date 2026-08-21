package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"strings"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
	"github.com/Halturshik/TicketAgregator-API/internal/documents/repository"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
)

const (
	StatusPending  = "pending"
	StatusVerified = "verified"
	StatusRejected = "rejected"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) documents.Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, ownerUserID int, passengerID *int) ([]documents.Document, error) {
	if passengerID != nil {
		return s.repo.ListForPassenger(ctx, ownerUserID, *passengerID)
	}
	return s.repo.ListForUser(ctx, ownerUserID)
}

func (s *Service) Create(ctx context.Context, ownerUserID int, in documents.SaveDocumentInput) (*documents.Document, error) {
	in.Type = strings.TrimSpace(in.Type)
	in.Number = strings.TrimSpace(in.Number)
	if !validator.ValidDocumentNumber(in.Type, in.Number) {
		return nil, apierror.ErrInvalidDocument
	}

	expiresAt, err := parseOptionalDate(in.ExpiresAt)
	if err != nil {
		return nil, apierror.ErrInvalidRequest
	}

	status, err := s.verify(in.Type)
	if err != nil {
		logger.Error("Ошибка mock-проверки документа userID=%d: %v", ownerUserID, err)
		return nil, err
	}
	var doc *documents.Document
	if in.PassengerID != nil {
		doc, err = s.repo.CreateForPassenger(ctx, ownerUserID, *in.PassengerID, in, status, expiresAt)
	} else {
		doc, err = s.repo.CreateForUser(ctx, ownerUserID, in, status, expiresAt)
	}
	if err == sql.ErrNoRows {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка сохранения документа userID=%d type=%s: %v", ownerUserID, in.Type, err)
		return nil, err
	}
	logger.Info("Документ сохранен: userID=%d documentID=%d status=%s", ownerUserID, doc.ID, doc.VerificationStatus)
	return doc, nil
}

func (s *Service) ValidateForBooking(ctx context.Context, in documents.BookingValidationInput) error {
	if in.DepartureTime.IsZero() || in.Transport == "" {
		return apierror.ErrInvalidRequest
	}

	useLatinNames := !in.Passenger.IsRussian || in.IsInternational
	if !validBookingNames(in.Passenger, useLatinNames) {
		return apierror.Validation(map[string]string{
			apierror.FieldPassengers: "Некорректно указаны данные пассажира",
		})
	}

	birthDate, ok := validator.ValidPassengerBirthDate(in.Passenger.BirthDate, in.DepartureTime)
	if !ok {
		return apierror.Validation(map[string]string{
			apierror.FieldBirthDate: "Некорректно указана дата рождения пассажира",
		})
	}

	documentType := strings.TrimSpace(in.Document.Type)
	if !validator.ValidDocumentNumber(documentType, in.Document.Number) {
		return apierror.ErrInvalidDocument
	}

	rule, err := s.repo.FindRule(ctx, in.Transport, in.IsInternational, validator.AgeOn(birthDate, in.DepartureTime), in.Passenger.IsRussian)
	if err == sql.ErrNoRows {
		logger.Warn("Не найдено правило документа transport=%s international=%t russian=%t", in.Transport, in.IsInternational, in.Passenger.IsRussian)
		return apierror.ErrDocumentNotAllowed
	}
	if err != nil {
		logger.Error("Ошибка поиска правила документа: %v", err)
		return err
	}
	if !rule.Allows(documentType) {
		return apierror.ErrDocumentNotAllowed
	}
	return nil
}

func validBookingNames(passenger documents.BookingPassenger, useLatinNames bool) bool {
	validName := validator.ValidCyrillicName
	if useLatinNames {
		validName = validator.ValidLatinName
	}
	if !validName(passenger.FirstName) || !validName(passenger.LastName) {
		return false
	}
	return strings.TrimSpace(passenger.MiddleName) == "" || validName(passenger.MiddleName)
}

func (s *Service) verify(documentType string) (string, error) {
	switch documentType {
	case "internal_passport", "international_passport":
		var value [1]byte
		if _, err := rand.Read(value[:]); err != nil {
			return "", err
		}
		if int(value[0])%100 < 5 {
			return StatusRejected, nil
		}
		return StatusVerified, nil
	case "birth_certificate":
		return StatusVerified, nil
	default:
		return StatusPending, nil
	}
}

func parseOptionalDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
