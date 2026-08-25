package documents

type PaymentDocument struct {
	SavedDocumentID    *int   `json:"-"`
	Type               string `json:"type"`
	Number             string `json:"number"`
	ExpiresAt          string `json:"expires_at,omitempty"`
	VerificationStatus string `json:"verification_status"`
	LastCheckedAt      string `json:"last_checked_at"`
	Fingerprint        string `json:"-"`
}
