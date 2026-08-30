package auth

import "errors"

var (
	ErrDuplicateEmail          = errors.New("duplicate email")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidRegistrationData = errors.New("invalid registration data")
)
