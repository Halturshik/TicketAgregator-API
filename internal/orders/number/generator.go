package number

import (
	"fmt"

	transportpkg "github.com/Halturshik/TicketAgregator-API/internal/transport"
)

type Series struct {
	transport string
	prefix    string
	issued    map[string]struct{}
}

func NewSeries(transport string) (*Series, error) {
	series := &Series{transport: transport, issued: make(map[string]struct{})}
	var err error
	switch transport {
	case transportpkg.Avia:
		series.prefix, err = randomAirPrefix()
	case transportpkg.Rail:
		series.prefix, err = randomRailPrefix()
	case transportpkg.Bus:
		series.prefix, err = randomBusSuffix()
	default:
		return nil, fmt.Errorf("unsupported transport %q", transport)
	}
	if err != nil {
		return nil, err
	}
	return series, nil
}

func (s *Series) Next() (string, error) {
	for range maxAttempts {
		number, err := s.next()
		if err != nil {
			return "", err
		}
		if _, exists := s.issued[number]; exists {
			continue
		}
		s.issued[number] = struct{}{}
		return number, nil
	}
	return "", fmt.Errorf("cannot generate unique ticket number for series")
}

func (s *Series) next() (string, error) {
	switch s.transport {
	case transportpkg.Rail:
		tail, err := randomRailTail()
		return s.prefix + tail, err
	case transportpkg.Bus:
		head, err := randomDigits(busHeadDigits)
		return head + s.prefix, err
	default:
		tail, err := randomDigits(airTailDigits)
		return s.prefix + tail, err
	}
}
