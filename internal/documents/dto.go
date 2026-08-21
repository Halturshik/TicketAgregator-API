package documents

import "time"

type Document struct {
	ID                 int    `json:"id"`
	OwnerUserID        *int   `json:"owner_user_id,omitempty"`
	PassengerID        *int   `json:"passenger_id,omitempty"`
	Type               string `json:"type"`
	Number             string `json:"number"`
	VerificationStatus string `json:"verification_status"`
	ExpiresAt          string `json:"expires_at,omitempty"`
}

type SaveDocumentInput struct {
	PassengerID *int   `json:"passenger_id,omitempty"`
	Type        string `json:"type"`
	Number      string `json:"number"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

type BookingPassenger struct {
	FirstName  string
	MiddleName string
	LastName   string
	BirthDate  string
	IsRussian  bool
}

type BookingDocument struct {
	Type   string
	Number string
}

type BookingValidationInput struct {
	Passenger       BookingPassenger
	Document        BookingDocument
	Transport       string
	IsInternational bool
	DepartureTime   time.Time
}

type DocumentRule struct {
	Transport                  string
	IsInternational            bool
	MinAge                     int
	MaxAge                     int
	IsRussian                  bool
	AllowInternalPassport      bool
	AllowInternationalPassport bool
	AllowBirthCertificate      bool
	AllowForeignPassport       bool
}

func (r DocumentRule) Allows(documentType string) bool {
	switch documentType {
	case "internal_passport":
		return r.AllowInternalPassport
	case "international_passport":
		return r.AllowInternationalPassport
	case "birth_certificate":
		return r.AllowBirthCertificate
	case "foreign_passport":
		return r.AllowForeignPassport
	default:
		return false
	}
}
