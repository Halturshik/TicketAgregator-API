package repository

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public cities: %w", err)
	}
	return result, nil
}
