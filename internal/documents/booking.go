package documents

import "time"

type BookingPassenger struct {
	FirstName  string
	MiddleName string
	LastName   string
	BirthDate  string
	IsRussian  bool
}

type BookingDocument struct {
	ID        *int
	Type      string
	Number    string
	ExpiresAt string
}

type BookingValidationInput struct {
	OwnerUserID      *int
	SavedPassengerID *int
	Passenger        BookingPassenger
	Document         BookingDocument
	Transport        string
	IsInternational  bool
	DepartureTime    time.Time
}

type ValidatedBookingDocument struct {
	SavedDocumentID    *int
	Type               string
	Number             string
	ExpiresAt          string
	VerificationStatus string
	LastCheckedAt      string
	Fingerprint        string
}
