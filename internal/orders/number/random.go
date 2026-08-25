package number

import "crypto/rand"

func randomDigits(length int) (string, error) {
	return randomCharacters(length, '0', digitAlphabetSize)
}

func randomLetters(length int) (string, error) {
	return randomCharacters(length, 'A', letterAlphabetSize)
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
