package trips

import "regexp"

var (
	aviaRouteNumberPattern = regexp.MustCompile(`^[A-Z0-9]{2,3} [0-9]{5}$`)
	railRouteNumberPattern = regexp.MustCompile(`^[A-Z0-9]{2,3} [0-9]{3}[A-Z]$`)
	busRouteNumberPattern  = regexp.MustCompile(`^[A-Z0-9]{2,3} [A-Z]{2}[0-9]{4}$`)
)

func ValidRouteNumber(value string) bool {
	return aviaRouteNumberPattern.MatchString(value) ||
		railRouteNumberPattern.MatchString(value) ||
		busRouteNumberPattern.MatchString(value)
}
