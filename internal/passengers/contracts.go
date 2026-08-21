package passengers

import "context"

type Service interface {
	List(ctx context.Context, ownerUserID int) ([]Passenger, error)
	Create(ctx context.Context, ownerUserID int, in SavePassengerInput) (*Passenger, error)
	Update(ctx context.Context, ownerUserID int, passengerID int, in SavePassengerInput) (*Passenger, error)
}
