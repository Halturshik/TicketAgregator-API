package repository

import "time"

type CreateUserParams struct {
	FirstName    string
	MiddleName   string
	LastName     string
	BirthDate    time.Time
	Email        string
	PasswordHash string
	IsRussian    bool
}

type UserAuth struct {
	ID           int
	Email        string
	PasswordHash string
	TokenVersion int
}
