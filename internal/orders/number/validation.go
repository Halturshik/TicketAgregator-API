package number

import "regexp"

var (
	airTicketPattern  = regexp.MustCompile(`^[A-Z]{2}-[0-9]{8}$`)
	railTicketPattern = regexp.MustCompile(`^[0-9][A-Z][0-9]{2}[A-Z][0-9][A-Z]$`)
	busTicketPattern  = regexp.MustCompile(`^[0-9]{6}[A-Z]$`)
	orderPattern      = regexp.MustCompile(`^[A-Z]{3}-[0-9]{5}$`)
)

func ValidTicket(value string) bool {
	return airTicketPattern.MatchString(value) || railTicketPattern.MatchString(value) ||
		busTicketPattern.MatchString(value)
}

func ValidOrder(value string) bool {
	return orderPattern.MatchString(value)
}
