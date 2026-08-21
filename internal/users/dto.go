package users

type Profile struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	MiddleName  string `json:"middle_name,omitempty"`
	LastName    string `json:"last_name"`
	BirthDate   string `json:"birth_date"`
	IsRussian   bool   `json:"is_russian"`
	BonusPoints int    `json:"bonus_points"`
}

type UpdateProfileInput struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName   string `json:"last_name"`
	BirthDate  string `json:"birth_date"`
	IsRussian  bool   `json:"is_russian"`
}
