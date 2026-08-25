package provider

import "math/rand"

func (p *MockProvider) price(rnd *rand.Rand, international bool) int {
	base := BasePriceByTransport[p.transport]
	if international {
		base = int(float64(base) * InternationalPriceMultiplier)
	}
	return base + rnd.Intn(base/PriceRandomDivisor)
}
