package documents

type Document struct {
	ID                 int    `json:"id"`
	OwnerUserID        *int   `json:"owner_user_id,omitempty"`
	PassengerID        *int   `json:"passenger_id,omitempty"`
	Type               string `json:"type"`
	Number             string `json:"number"`
	VerificationStatus string `json:"verification_status"`
	ExpiresAt          string `json:"expires_at,omitempty"`
	LastCheckedAt      string `json:"last_checked_at,omitempty"`
	Fingerprint        string `json:"-"`
}
