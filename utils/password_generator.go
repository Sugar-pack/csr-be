package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	MinPasswordLen                    = 4
	MaxPasswordLen                    = 32
	AllowedRandomResetPasswordSymbols = "abcdefghijklmnopqrstuvwxyz"
	GeneratedRandomResetPasswordLen   = 8
)

var UnsupportedPasswordLengthErr = fmt.Errorf(
	"length should be within the interval [%d : %d]",
	MinPasswordLen,
	MaxPasswordLen)

func GenerateRandomResetPassword() (string, error) {
	return generateRandomPassword(
		GeneratedRandomResetPasswordLen,
		AllowedRandomResetPasswordSymbols)
}

func generateRandomString(n int, symbols string) (string, error) {
	if n < MinPasswordLen || n > MaxPasswordLen {
		return "", UnsupportedPasswordLengthErr
	}

	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(symbols))))
		if err != nil {
			return "", err
		}
		ret[i] = symbols[num.Int64()]
	}

	return string(ret), nil
}

func generateRandomPassword(length int, symbols string) (string, error) {
	return generateRandomString(length, symbols)
}
