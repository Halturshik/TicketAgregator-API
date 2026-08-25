package documents

type SaveDocumentInput struct {
	PassengerID *int   `json:"passenger_id,omitempty"`
	Type        string `json:"type"`
	Number      string `json:"number"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}
