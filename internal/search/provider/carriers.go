package provider

import "math/rand"

type Carrier struct {
	ID   int
	Name string
	Code string
}

func (p *MockProvider) pickCarrier(rnd *rand.Rand) Carrier {
	return p.carriers[rnd.Intn(len(p.carriers))]
}

func (p *MockProvider) pickSecondCarrier(rnd *rand.Rand, first Carrier) Carrier {
	if rnd.Intn(ProbabilityScale) >= CarrierChangeOnTransferPercent {
		return first
	}
	if len(p.carriers) < 2 {
		return first
	}

	for {
		candidate := p.pickCarrier(rnd)
		if candidate.ID != first.ID || candidate.Code != first.Code {
			return candidate
		}
	}
}

func (p *MockProvider) pickDifferentCarrier(rnd *rand.Rand, first Carrier) Carrier {
	if len(p.carriers) < 2 {
		return first
	}
	for {
		candidate := p.pickCarrier(rnd)
		if candidate.ID != first.ID || candidate.Code != first.Code {
			return candidate
		}
	}
}
