package provider

import (
	"math"
	"math/rand"
	"sort"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (p *MockProvider) pickHub(rnd *rand.Rand, from search.City, to search.City, cities []search.City) (search.City, bool) {
	type candidate struct {
		city  search.City
		score float64
	}
	candidates := []candidate{}
	for _, city := range cities {
		if city.ID == from.ID || city.ID == to.ID || !city.IsAirHub {
			continue
		}
		candidates = append(candidates, candidate{
			city:  city,
			score: routeDistance(from, city) + routeDistance(city, to),
		})
	}
	if len(candidates) == 0 {
		return search.City{}, false
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score < candidates[j].score
	})

	bestCount := min(HubCandidatesLimit, len(candidates))
	return candidates[rnd.Intn(bestCount)].city, true
}

func routeDistance(from search.City, to search.City) float64 {
	if from.Latitude == 0 && from.Longitude == 0 || to.Latitude == 0 && to.Longitude == 0 {
		if from.CountryID == to.CountryID {
			return SameCountryFallbackDistance
		}
		return CrossCountryFallbackDistance
	}
	return math.Hypot(from.Latitude-to.Latitude, from.Longitude-to.Longitude)
}
