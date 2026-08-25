package passengers

import (
	"github.com/Halturshik/TicketAgregator-API/internal/documents"
)

type PaymentSnapshot struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName   string `json:"last_name"`
	BirthDate  string `json:"birth_date"`
	IsRussian  bool   `json:"is_russian"`
}

type PaymentPassenger struct {
	Source           string
	SavedPassengerID *int
	SaveChanges      bool
	Passenger        PaymentSnapshot
	Document         documents.PaymentDocument
}
