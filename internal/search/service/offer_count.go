package service

import "time"

type offerCountRange struct {
	base     int
	maxExtra int
}

func generationRange(searchDate time.Time, today time.Time) offerCountRange {
	switch {
	case !searchDate.After(today.AddDate(0, FirstHorizonMonths, 0)):
		return offerCountRange{base: FirstHorizonBase, maxExtra: FirstHorizonExtra}
	case !searchDate.After(today.AddDate(0, SecondHorizonMonths, 0)):
		return offerCountRange{base: SecondHorizonBase, maxExtra: SecondHorizonExtra}
	case !searchDate.After(today.AddDate(0, ThirdHorizonMonths, 0)):
		return offerCountRange{base: ThirdHorizonBase, maxExtra: ThirdHorizonExtra}
	default:
		return offerCountRange{base: FourthHorizonBase, maxExtra: FourthHorizonExtra}
	}
}

func generatedOffersCount(searchDate time.Time, today time.Time, entropy int64) int {
	rule := generationRange(searchDate, today)
	if entropy < 0 {
		entropy = -entropy
	}
	return rule.base + int(entropy%int64(rule.maxExtra+1))
}
