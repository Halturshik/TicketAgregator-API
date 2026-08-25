package payments

type Record struct {
	OrderID  int
	Amount   int
	Status   string
	Provider string
}
