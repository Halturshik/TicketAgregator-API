package types

type RegisterInput struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName   string `json:"last_name"`
	BirthDate  string `json:"birth_date"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	IsRussian  bool   `json:"is_russian"`
}

type ConfirmRegisterInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
