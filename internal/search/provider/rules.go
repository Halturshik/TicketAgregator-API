package provider

import (
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/transport"
)

const (
	ProbabilityScale = 100
	RandomSeedStride = 7919

	// Шанс пересадки (только avia + international), в процентах
	TransferChancePercent = 35

	// С какой вероятностью при пересадке меняется перевозчик на втором сегменте
	CarrierChangeOnTransferPercent = 15
	SameCarrierOnReturnPercent     = 75

	// Сколько ближайших хабов рассматриваем при выборе пересадки
	HubCandidatesLimit = 3

	// Множитель цены для международных рейсов
	InternationalPriceMultiplier = 1.7

	// ----- Длительности (минуты) -----

	// Пересадка: первый сегмент
	TransferFirstLegMinMinutes = 120
	TransferFirstLegMaxExtra   = 180

	// Ожидание на пересадке
	TransferWaitMinMinutes = 60
	TransferWaitMaxExtra   = 120

	// Второй сегмент после пересадки
	TransferSecondLegMinMinutes = 120
	TransferSecondLegMaxExtra   = 240

	// Прямые рейсы
	AviaDomesticMinMinutes = 90
	AviaDomesticMaxExtra   = 180

	AviaIntlMinMinutes = 180
	AviaIntlMaxExtra   = 300

	RailMinMinutes = 240
	RailMaxExtra   = 600

	BusMinMinutes = 120
	BusMaxExtra   = 360

	DepartureMinuteStep  = 10
	MinutesPerDay        = 24 * 60
	ReturnConnectionTime = time.Hour
	PriceRandomDivisor   = 2

	RailRouteNumberRange = 1_000
	BusRouteNumberRange  = 10_000
	AviaRouteNumberRange = 100_000
	LatinAlphabetSize    = 26
	DefaultCarrierCode   = "MCK"

	SameCountryFallbackDistance  = 1.0
	CrossCountryFallbackDistance = 100.0
)

// Базовые цены до рандома и умножения на пассажиров
var BasePriceByTransport = map[string]int{
	transport.Avia: 9000,
	transport.Rail: 3500,
	transport.Bus:  1800,
}
