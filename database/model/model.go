package model

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

type User struct {
	ID           int
	FirstName    string
	MiddleName   string
	LastName     string
	BirthDate    time.Time
	Email        string
	PasswordHash string
	IsRussian    bool
}
