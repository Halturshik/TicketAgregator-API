package documents

import "errors"

var (
	ErrNotFound      = errors.New("document not found")
	ErrAlreadyExists = errors.New("document already exists")
	ErrRuleNotFound  = errors.New("document rule not found")
)
