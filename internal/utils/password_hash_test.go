package utils

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHash(t *testing.T) {
	password := "password"
	hashedPassword, err := PasswordHash(password)
	if err != nil {
		t.Errorf("error while hashing password: %s", err)
	}
	if len(hashedPassword) == 0 {
		t.Error("hashed password is empty")
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		t.Errorf("error while comparing password hash: %s", err)
	}
}
