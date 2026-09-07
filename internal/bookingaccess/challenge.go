package bookingaccess

type Challenge struct {
	ID          string `json:"id"`
	OrderID     int    `json:"order_id"`
	OrderNumber string `json:"order_number"`
	Locator     string `json:"locator"`
	Email       string `json:"email"`
}
