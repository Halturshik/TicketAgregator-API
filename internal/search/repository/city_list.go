package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

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
		city, err := scanCity(rows)
		if err != nil {
			return nil, fmt.Errorf("scan city: %w", err)
		}
		result = append(result, *city)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cities: %w", err)
	}
	return result, nil
}
