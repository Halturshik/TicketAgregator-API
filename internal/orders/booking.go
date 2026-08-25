package orders

type DocumentBooking struct {
	SavedDocumentID *int   `json:"saved_document_id,omitempty"`
	Type            string `json:"type,omitempty"`
	Number          string `json:"number,omitempty"`
	ExpiresAt       string `json:"expires_at,omitempty"`
}

type PassengerBooking struct {
	Source           string            `json:"source"`
	SavedPassengerID *int              `json:"saved_passenger_id,omitempty"`
	SaveChanges      bool              `json:"save_changes,omitempty"`
	Passenger        PassengerSnapshot `json:"passenger"`
	Document         DocumentBooking   `json:"document"`
}

type CreateOrderInput struct {
	SearchID     string             `json:"search_id"`
	TripOptionID string             `json:"trip_option_id"`
	GuestEmail   string             `json:"guest_email,omitempty"`
	UseBonus     int                `json:"use_bonus,omitempty"`
	Passengers   []PassengerBooking `json:"passengers"`
}
