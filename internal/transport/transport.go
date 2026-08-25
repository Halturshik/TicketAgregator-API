package transport

const (
	Avia = "avia"
	Rail = "rail"
	Bus  = "bus"
)

func IsSupported(value string) bool {
	return value == Avia || value == Rail || value == Bus
}

func IsGround(value string) bool {
	return value == Rail || value == Bus
}
