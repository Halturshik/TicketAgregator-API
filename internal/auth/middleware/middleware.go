package middleware

import (
	"github.com/Halturshik/TicketAgregator-API/internal/auth"
)

type Middleware struct {
	users  auth.UserReader
	tokens auth.AccessTokenParser
}

func New(users auth.UserReader, tokens auth.AccessTokenParser) *Middleware {
	return &Middleware{
		users:  users,
		tokens: tokens,
	}
}
