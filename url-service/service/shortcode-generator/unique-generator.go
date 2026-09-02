package shortcodegenerator

import (
	"crypto/rand"
	"errors"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type UniqueGenerator struct{}

func NewUniqueGenerator() *UniqueGenerator {
	return &UniqueGenerator{}
}

func (g *UniqueGenerator) GenerateUnique(
	sequence int64,
	randomLength int,
) (string, error) {

	if sequence <= 0 {
		return "", errors.New("invalid sequence")
	}

	if randomLength <= 0 {
		return "", errors.New("invalid random length")
	}

	counterPart := encodeBase62(sequence)

	randomPart, err := generateRandomLetters(randomLength)
	if err != nil {
		return "", err
	}

	return counterPart + randomPart, nil
}

func encodeBase62(n int64) string {
	if n == 0 {
		return string(base62Alphabet[0])
	}

	var digits []byte
	for n > 0 {
		digits = append(digits, base62Alphabet[n%62])
		n /= 62
	}

	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	return string(digits)
}

func generateRandomLetters(length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := range bytes {
		bytes[i] = base62Alphabet[int(bytes[i])%62]
	}

	return string(bytes), nil
}