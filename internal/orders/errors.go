package orders

import "errors"

var (
	ErrNotFound     = errors.New("order not found")
	ErrCannotBePaid = errors.New("order cannot be paid")
	ErrExpired      = errors.New("order expired")
)
