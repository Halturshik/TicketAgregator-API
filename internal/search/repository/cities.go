package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/lib/pq"
)

func (r *Repository) GetCity(ctx context.Context, id int) (*search.City, error) {
	var c search.City
	err := r.DB.QueryRowContext(ctx, `
		SELECT c.id, c.name, c.country_id, co.name, co.is_russia, c.is_air_hub,
		       co.neighbor_country_ids,
		       COALESCE(c.latitude, 0), COALESCE(c.longitude, 0)
		FROM cities c
		JOIN countries co ON co.id = c.country_id
		WHERE c.id = $1
	`, id).Scan(&c.ID, &c.Name, &c.CountryID, &c.Country, &c.IsRussia, &c.IsAirHub, pq.Array(&c.NeighborCountryIDs), &c.Latitude, &c.Longitude)
	if err != nil {
		return nil, fmt.Errorf("get city: %w", err)
	}
	return &c, nil
}

func (r *Repository) ListCities(ctx context.Context) ([]search.City, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT c.id, c.name, c.country_id, co.name, co.is_russia, c.is_air_hub,
		       co.neighbor_country_ids,
		       COALESCE(c.latitude, 0), COALESCE(c.longitude, 0)
		FROM cities c
		JOIN countries co ON co.id = c.country_id
		ORDER BY c.id
	`)
	if err != nil {
		return nil, fmt.Errorf("list cities: %w", err)
	}
	defer rows.Close()

	result := []search.City{}
	for rows.Next() {
		var c search.City
		if err := rows.Scan(&c.ID, &c.Name, &c.CountryID, &c.Country, &c.IsRussia, &c.IsAirHub, pq.Array(&c.NeighborCountryIDs), &c.Latitude, &c.Longitude); err != nil {
			return nil, fmt.Errorf("scan city: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *Repository) ListCitiesPublic(ctx context.Context) ([]search.CityPublic, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT c.id, c.name, co.name, c.is_air_hub
		FROM cities c
		JOIN countries co ON co.id = c.country_id
		ORDER BY co.name, c.name
	`)
	if err != nil {
		return nil, fmt.Errorf("list cities public: %w", err)
	}
	defer rows.Close()

	result := make([]search.CityPublic, 0)
	for rows.Next() {
		var item search.CityPublic
		if err := rows.Scan(&item.ID, &item.Name, &item.Country, &item.IsHub); err != nil {
			return nil, fmt.Errorf("scan city public: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
