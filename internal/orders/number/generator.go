package number

import (
	"crypto/rand"
	"fmt"
)

const maxAttempts = 20

type Series struct {
	transport string
	prefix    string
	issued    map[string]struct{}
}

func NewSeries(transport string) (*Series, error) {
	series := &Series{transport: transport, issued: make(map[string]struct{})}
	var err error
	switch transport {
	case "rail":
		series.prefix, err = randomDigitLetterDigit()
	case "bus":
		series.prefix, err = randomDigitLetter()
	default:
		series.prefix, err = randomAirPrefix()
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
	case "rail":
		tail, err := randomRailTail()
		return s.prefix + tail, err
	case "bus":
		head, err := randomDigits(5)
		return head + s.prefix, err
	default:
		tail, err := randomDigits(5)
		return s.prefix + tail, err
	}
}

func randomAirPrefix() (string, error) {
	letters, err := randomLetters(2)
	if err != nil {
		return "", err
	}
	digits, err := randomDigits(3)
	if err != nil {
		return "", err
	}
	return letters + "-" + digits, nil
}

func randomDigitLetterDigit() (string, error) {
	first, err := randomDigits(1)
	if err != nil {
		return "", err
	}
	letter, err := randomLetters(1)
	if err != nil {
		return "", err
	}
	last, err := randomDigits(1)
	if err != nil {
		return "", err
	}
	return first + letter + last, nil
}

func randomRailTail() (string, error) {
	digits, err := randomDigits(2)
	if err != nil {
		return "", err
	}
	firstLetter, err := randomLetters(1)
	if err != nil {
		return "", err
	}
	digit, err := randomDigits(1)
	if err != nil {
		return "", err
	}
	lastLetter, err := randomLetters(1)
	if err != nil {
		return "", err
	}
	return digits + firstLetter + digit + lastLetter, nil
}

func randomDigitLetter() (string, error) {
	digit, err := randomDigits(1)
	if err != nil {
		return "", err
	}
	letter, err := randomLetters(1)
	if err != nil {
		return "", err
	}
	return digit + letter, nil
}

func randomDigits(length int) (string, error) {
	return randomCharacters(length, '0', 10)
}

func randomLetters(length int) (string, error) {
	return randomCharacters(length, 'A', 26)
}

func randomCharacters(length int, first byte, alphabetSize byte) (string, error) {
	value := make([]byte, length)
	for index := range value {
		var source [1]byte
		if _, err := rand.Read(source[:]); err != nil {
			return "", err
		}
		value[index] = first + source[0]%alphabetSize
	}
	return string(value), nil
}
