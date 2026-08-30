package auth

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

type PendingRegistration struct {
	FirstName    string `json:"first_name"`
	MiddleName   string `json:"middle_name,omitempty"`
	LastName     string `json:"last_name"`
	BirthDate    string `json:"birth_date"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	IsRussian    bool   `json:"is_russian"`
}

type UserAuth struct {
	ID           int
	Email        string
	PasswordHash string
	TokenVersion int
}
