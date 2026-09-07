package codegen

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const DefaultNumericLength = 4

const numericAlphabetSize = 10

type NumericGenerator struct {
	length int
}

func NewNumericGenerator(length int) *NumericGenerator {
	if length <= 0 {
		length = DefaultNumericLength
	}
	return &NumericGenerator{length: length}
}

func (g *NumericGenerator) GenerateVerificationCode() (string, error) {
	code := make([]byte, g.length)
	for index := range code {
		value, err := rand.Int(rand.Reader, big.NewInt(numericAlphabetSize))
		if err != nil {
			return "", fmt.Errorf("generate verification code digit: %w", err)
		}
		code[index] = byte('0' + value.Int64())
	}
	return string(code), nil
}
