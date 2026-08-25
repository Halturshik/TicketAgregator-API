package orders

const (
	DefaultListLimit = 10
	MaxListLimit     = 50
)

type ListFilter struct {
	Transport string
	Limit     int
	Offset    int
}

func NormalizeListLimit(limit int) int {
	if limit <= 0 || limit > MaxListLimit {
		return DefaultListLimit
	}
	return limit
}
