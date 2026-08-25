package provider

import (
	"math/rand"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/transport"
)

func randomDeparture(date time.Time, earliest time.Time, rnd *rand.Rand) time.Time {
	slotsPerDay := MinutesPerDay / DepartureMinuteStep
	firstSlot := 0
	if !earliest.IsZero() && !earliest.Before(date) {
		minutes := int(earliest.Sub(date).Minutes())
		firstSlot = (minutes + DepartureMinuteStep - 1) / DepartureMinuteStep
		if firstSlot >= slotsPerDay {
			return date.Add(time.Duration(firstSlot*DepartureMinuteStep) * time.Minute)
		}
	}
	slot := firstSlot + rnd.Intn(slotsPerDay-firstSlot)
	return date.Add(time.Duration(slot*DepartureMinuteStep) * time.Minute)
}

func (p *MockProvider) duration(rnd *rand.Rand, international bool) int {
	switch p.transport {
	case transport.Rail:
		return RailMinMinutes + rnd.Intn(RailMaxExtra)
	case transport.Bus:
		return BusMinMinutes + rnd.Intn(BusMaxExtra)
	default:
		if international {
			return AviaIntlMinMinutes + rnd.Intn(AviaIntlMaxExtra)
		}
		return AviaDomesticMinMinutes + rnd.Intn(AviaDomesticMaxExtra)
	}
}
