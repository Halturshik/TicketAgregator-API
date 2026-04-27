package code

import (
	"crypto/rand"
	"math/big"
)

const numbers = "0123456789"

func GenerateVerificationCode() (string, error) {
	code := make([]byte, 4)

	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(numbers))))
		if err != nil {
			return "", err
		}
		code[i] = numbers[num.Int64()]
	}

	return string(code), nil
}
