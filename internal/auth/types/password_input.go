package types

type ChangePasswordStartInput struct {
	Email string `json:"email"`
}

type PasswordVerifyInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type ChangePasswordConfirmInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
