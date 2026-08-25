package provider

import (
	"fmt"
	"math/rand"
	"strings"

	transportpkg "github.com/Halturshik/TicketAgregator-API/internal/transport"
)

func routeNumber(transport string, carrierCode string, rnd *rand.Rand) string {
	code := strings.ToUpper(strings.TrimSpace(carrierCode))
	if code == "" {
		code = DefaultCarrierCode
	}

	switch transport {
	case transportpkg.Rail:
		return fmt.Sprintf("%s %03d%c", code, rnd.Intn(RailRouteNumberRange), latinLetter(rnd))
	case transportpkg.Bus:
		return fmt.Sprintf("%s %c%c%04d", code, latinLetter(rnd), latinLetter(rnd), rnd.Intn(BusRouteNumberRange))
	default:
		return fmt.Sprintf("%s %05d", code, rnd.Intn(AviaRouteNumberRange))
	}
}

func latinLetter(rnd *rand.Rand) byte {
	return byte('A' + rnd.Intn(LatinAlphabetSize))
}
