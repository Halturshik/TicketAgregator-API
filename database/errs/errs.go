package errs

import "errors"

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)
