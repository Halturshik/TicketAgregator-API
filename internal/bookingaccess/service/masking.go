package service

import (
	"fmt"
	"strings"
)

func maskedPassenger(firstName string, lastName string) string {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	for _, initial := range []rune(lastName) {
		return fmt.Sprintf("%s %c.", firstName, initial)
	}
	return firstName
}

func maskedDocument(number string) string {
	runes := []rune(strings.TrimSpace(number))
	if len(runes) > 4 {
		runes = runes[len(runes)-4:]
	}
	return "**** " + string(runes)
}
