package bonus

const (
	DefaultListLimit = 10
	MaxListLimit     = 50
)

func NormalizeListLimit(limit int) int {
	if limit <= 0 || limit > MaxListLimit {
		return DefaultListLimit
	}
	return limit
}
