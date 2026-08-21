package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/passengers"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func scanPassenger(scanner interface {
	Scan(dest ...any) error
}) (*passengers.Passenger, error) {
	var p passengers.Passenger
	var birthDate time.Time
	if err := scanner.Scan(&p.ID, &p.FirstName, &p.MiddleName, &p.LastName, &birthDate, &p.IsRussian); err != nil {
		return nil, err
	}
	p.BirthDate = birthDate.Format("2006-01-02")
	return &p, nil
}

func (r *Repository) List(ctx context.Context, ownerUserID int) ([]passengers.Passenger, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, first_name, COALESCE(middle_name, ''), last_name, birth_date, is_russian
		FROM saved_passengers
		WHERE owner_user_id = $1
		ORDER BY id DESC
	`, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("list passengers: %w", err)
	}
	defer rows.Close()

	result := []passengers.Passenger{}
	for rows.Next() {
		p, err := scanPassenger(rows)
		if err != nil {
			return nil, fmt.Errorf("scan passenger: %w", err)
		}
		result = append(result, *p)
	}
	return result, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerUserID int, in passengers.SavePassengerInput, birthDate time.Time) (*passengers.Passenger, error) {
	return scanPassenger(r.DB.QueryRowContext(ctx, `
		INSERT INTO saved_passengers (owner_user_id, first_name, middle_name, last_name, birth_date, is_russian)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, first_name, COALESCE(middle_name, ''), last_name, birth_date, is_russian
	`, ownerUserID, in.FirstName, in.MiddleName, in.LastName, birthDate, in.IsRussian))
}

func (r *Repository) Update(ctx context.Context, ownerUserID int, passengerID int, in passengers.SavePassengerInput, birthDate time.Time) (*passengers.Passenger, error) {
	return scanPassenger(r.DB.QueryRowContext(ctx, `
		UPDATE saved_passengers
		SET first_name = $1,
		    middle_name = $2,
		    last_name = $3,
		    birth_date = $4,
		    is_russian = $5,
		    updated_at = NOW()
		WHERE id = $6 AND owner_user_id = $7
		RETURNING id, first_name, COALESCE(middle_name, ''), last_name, birth_date, is_russian
	`, in.FirstName, in.MiddleName, in.LastName, birthDate, in.IsRussian, passengerID, ownerUserID))
}
