package validator

import "net/mail"

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
