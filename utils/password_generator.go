package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	MinPassLen = 4
	MaxPassLen = 32
)

var UnsupportedPasswordLengthErr = fmt.Errorf(
	"length should be within the interval [%d : %d]",
	MinPassLen,
	MaxPassLen)

func generateRandomString(n int, symbols string) (string, error) {
	if n < MinPassLen || n > MaxPassLen {
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

func GenerateRandomResetPassword() (string, error) {
	symbols := "abcdefghijklmnopqrstuvwxyz"

	return generateRandomResetPassword(8, symbols)
}

func generateRandomResetPassword(length int, symbols string) (string, error) {
	return generateRandomString(length, symbols)
}
