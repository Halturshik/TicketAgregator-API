package service

import (
	"strings"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/validator"
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

type preparedDocument struct {
	status      string
	fingerprint string
	expiresAt   *time.Time
	checkedAt   time.Time
}

func (s *Service) prepareDocument(in *documents.SaveDocumentInput) (*preparedDocument, error) {
	in.Type = strings.TrimSpace(in.Type)
	in.Number = strings.ToUpper(strings.TrimSpace(in.Number))
	if !validator.ValidDocumentNumber(in.Type, in.Number) {
		return nil, apierror.ErrInvalidDocument
	}
	expiresAt, err := parseOptionalDate(in.ExpiresAt)
	if err != nil {
		return nil, apierror.ErrInvalidRequest
	}
	checkedAt := s.now().UTC()
	status, fingerprint := verifyDocument(s.verificationSecret, in.Type, in.Number, checkedAt)
	return &preparedDocument{
		status: status, fingerprint: fingerprint, expiresAt: expiresAt, checkedAt: checkedAt,
	}, nil
}

func parseOptionalDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(documents.DateLayout, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
