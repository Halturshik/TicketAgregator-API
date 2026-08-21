package bonus

type Transaction struct {
	ID        int    `json:"id"`
	OrderID   *int   `json:"order_id,omitempty"`
	Type      string `json:"type"`
	Amount    int    `json:"amount"`
	CreatedAt string `json:"created_at"`
}
