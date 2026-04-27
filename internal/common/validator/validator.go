package validator

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

func NotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

func ValidName(s string) bool {
	s = strings.TrimSpace(s)
	nameRegex := regexp.MustCompile(`^[А-Яа-яЁё -]+$`)
	return nameRegex.MatchString(s)
}

func ValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidBirthDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}

	now := time.Now()
	minAgeDate := now.AddDate(-14, 0, 0)
	if t.After(minAgeDate) {
		return time.Time{}, errors.New("слишком молод")
	}

	return t, nil
}

func ValidPassword(p string) bool {
	if len(p) < 8 {
		return false
	}

	var hasLetter, hasNumber bool
	for _, c := range p {
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z':
			hasLetter = true
		case '0' <= c && c <= '9':
			hasNumber = true
		}
	}
	return hasLetter && hasNumber
}
