package types

type LoginStartInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginConfirmInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type LoginOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
