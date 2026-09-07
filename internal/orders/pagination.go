package orders

const (
	DefaultListLimit = 5
	MaxListLimit     = 5
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
