package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (r *Repository) GetCity(ctx context.Context, id int) (*search.City, error) {
	city, err := scanCity(r.DB.QueryRowContext(ctx, `
		SELECT c.id, c.name, c.country_id, co.name, co.is_russia, c.is_air_hub,
		       co.neighbor_country_ids,
		       COALESCE(c.latitude, 0), COALESCE(c.longitude, 0)
		FROM cities c
		JOIN countries co ON co.id = c.country_id
		WHERE c.id = $1
	`, id))
	if err == sql.ErrNoRows {
		return nil, search.ErrCityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get city: %w", err)
	}
	return city, nil
}
