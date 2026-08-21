package provider

const (
	WorkerLimit = 4

	// Шанс пересадки (только avia + international), в процентах
	TransferChancePercent = 35

	// С какой вероятностью при пересадке меняется перевозчик на втором сегменте
	CarrierChangeOnTransferPercent = 15

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

	DepartureMinuteStep = 10

	// Бонусы
	BonusRate = 0.02 // 2%
)

// Базовые цены до рандома и умножения на пассажиров
var BasePriceByTransport = map[string]int{
	"avia": 9000,
	"rail": 3500,
	"bus":  1800,
}
