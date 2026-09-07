package orders

type HistoryPage struct {
	Total  int            `json:"total"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
	Items  []HistoryOrder `json:"items"`
}

type HistoryOrder struct {
	ID                int                `json:"id"`
	OrderNumber       string             `json:"order_number"`
	Status            string             `json:"status"`
	CurrentTotalPrice int                `json:"current_total_price"`
	BonusSpent        int                `json:"bonus_spent"`
	BonusEarned       int                `json:"bonus_earned"`
	CreatedAt         string             `json:"created_at"`
	Directions        []HistoryDirection `json:"directions"`
	Tickets           []HistoryTicket    `json:"tickets"`
}

type HistoryTicket struct {
	ID           int              `json:"id"`
	TicketNumber string           `json:"ticket_number"`
	Status       string           `json:"status"`
	Transport    string           `json:"transport"`
	Passenger    HistoryPassenger `json:"passenger"`
	Document     HistoryDocument  `json:"document"`
	Direction    HistoryDirection `json:"direction"`
}

type HistoryPassenger struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName   string `json:"last_name"`
}

type HistoryDocument struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type HistoryDirection struct {
	FromCity      string `json:"from_city"`
	ToCity        string `json:"to_city"`
	DepartureTime string `json:"departure_time"`
	ArrivalTime   string `json:"arrival_time"`
}
