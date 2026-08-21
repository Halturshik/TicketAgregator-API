package provider

import (
	"fmt"
	"math/rand"
	"strings"
)

func routeNumber(transport string, carrierCode string, rnd *rand.Rand) string {
	code := strings.ToUpper(strings.TrimSpace(carrierCode))
	if code == "" {
		code = "MCK"
	}

	switch transport {
	case "rail":
		return fmt.Sprintf("%s %03d%c", code, rnd.Intn(1000), latinLetter(rnd))
	case "bus":
		return fmt.Sprintf("%s %c%c%04d", code, latinLetter(rnd), latinLetter(rnd), rnd.Intn(10000))
	default:
		return fmt.Sprintf("%s %05d", code, rnd.Intn(100000))
	}
}

func latinLetter(rnd *rand.Rand) byte {
	return byte('A' + rnd.Intn(26))
}
