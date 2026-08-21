package validator

import (
	"regexp"
	"strings"
	"time"
)

var (
	cyrillicNameRegex = regexp.MustCompile(`^[А-Яа-яЁё -]+$`)
	latinNameRegex    = regexp.MustCompile(`^[A-Za-z -]+$`)
	internalPassport  = regexp.MustCompile(`^\d{4}\s?\d{6}$`)
	internationalPass = regexp.MustCompile(`^[A-Z0-9]{6,12}$`)
	foreignPassport   = regexp.MustCompile(`^[A-Z0-9]{5,20}$`)
	birthCertificate  = regexp.MustCompile(`^[А-ЯA-Z0-9-]{5,20}$`)
)

func ValidCyrillicName(value string) bool {
	return cyrillicNameRegex.MatchString(strings.TrimSpace(value))
}

func ValidLatinName(value string) bool {
	return latinNameRegex.MatchString(strings.TrimSpace(value))
}

func ValidPassengerBirthDate(value string, departure time.Time) (time.Time, bool) {
	birthDate, err := time.Parse("2006-01-02", value)
	if err != nil || birthDate.After(departure) {
		return time.Time{}, false
	}
	return birthDate, true
}

func AgeOn(birthDate time.Time, date time.Time) int {
	age := date.Year() - birthDate.Year()
	if date.Month() < birthDate.Month() || (date.Month() == birthDate.Month() && date.Day() < birthDate.Day()) {
		age--
	}
	return age
}

func ValidDocumentNumber(documentType string, number string) bool {
	number = strings.ToUpper(strings.TrimSpace(number))
	switch documentType {
	case "internal_passport":
		return internalPassport.MatchString(number)
	case "international_passport":
		return internationalPass.MatchString(number)
	case "foreign_passport":
		return foreignPassport.MatchString(number)
	case "birth_certificate":
		return birthCertificate.MatchString(number)
	default:
		return false
	}
}
