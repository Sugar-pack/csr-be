package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func PasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("create user error, failed to generate password hash: %s", err)
	}
	return string(hashedPassword), nil
}
