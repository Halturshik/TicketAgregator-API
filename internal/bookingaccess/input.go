package bookingaccess

type Locator struct {
	TicketNumber string
	OrderNumber  string
}

type LookupInput struct {
	TicketNumber string `json:"ticket_number,omitempty"`
	OrderNumber  string `json:"order_number,omitempty"`
}

type AccessRequestInput struct {
	TicketNumber string `json:"ticket_number,omitempty"`
	OrderNumber  string `json:"order_number,omitempty"`
	Email        string `json:"email"`
}

type AccessConfirmInput struct {
	ChallengeID string `json:"challenge_id"`
	Code        string `json:"code"`
}

type RefundInput struct {
	TicketNumbers []string `json:"ticket_numbers,omitempty"`
	All           bool     `json:"all,omitempty"`
}
