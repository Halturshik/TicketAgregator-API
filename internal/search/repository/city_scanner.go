package repository

import (
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/lib/pq"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanCity(source scanner) (*search.City, error) {
	var city search.City
	var neighborIDs pq.Int64Array
	if err := source.Scan(
		&city.ID, &city.Name, &city.CountryID, &city.Country, &city.IsRussia, &city.IsAirHub,
		&neighborIDs, &city.Latitude, &city.Longitude,
	); err != nil {
		return nil, err
	}
	city.NeighborCountryIDs = intIDs(neighborIDs)
	return &city, nil
}

func intIDs(values []int64) []int {
	result := make([]int, len(values))
	for index, value := range values {
		result[index] = int(value)
	}
	return result
}
