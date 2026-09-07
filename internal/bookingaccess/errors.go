package bookingaccess

import "errors"

var (
	ErrNotFound     = errors.New("booking not found")
	ErrForbidden    = errors.New("booking access forbidden")
	ErrInvalidToken = errors.New("invalid booking access token")
)
