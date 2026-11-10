package users

import (
	"crypto/rand"
	"math/big"
)

const (
	generatedPasswordLength = 12
	lowercaseChars          = "abcdefghijklmnopqrstuvwxyz"
	uppercaseChars          = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars              = "0123456789"
	specialChars            = "!@#$%&*()_+=-[]{}|;:,.<>?"
)

func randomIndex(limit int) (int, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(int64(limit)))
	if err != nil {
		return 0, err
	}

	return int(number.Int64()), nil
}

func randomChar(chars string) (byte, error) {
	index, err := randomIndex(len(chars))
	if err != nil {
		return 0, err
	}

	return chars[index], nil
}

func generateRandomPassword() (string, error) {
	requiredGroups := []string{lowercaseChars, uppercaseChars, digitChars, specialChars}
	allChars := lowercaseChars + uppercaseChars + digitChars + specialChars

	password := make([]byte, 0, generatedPasswordLength)

	for _, group := range requiredGroups {
		char, err := randomChar(group)
		if err != nil {
			return "", err
		}

		password = append(password, char)
	}

	for len(password) < generatedPasswordLength {
		char, err := randomChar(allChars)
		if err != nil {
			return "", err
		}

		password = append(password, char)
	}

	for i := len(password) - 1; i > 0; i-- {
		j, err := randomIndex(i + 1)
		if err != nil {
			return "", err
		}

		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}
