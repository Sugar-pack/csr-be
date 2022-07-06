package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

type PasswordService struct {
	PasswordTokenTTL *time.Duration
	PasswordLength   int
}

func NewPasswordService() (*PasswordService, error) {
	passwordResetExpirationMinutes := os.Getenv("PASSWORD_RESET_EXPIRATION_MINUTES")
	if passwordResetExpirationMinutes == "" {
		return nil, fmt.Errorf("PASSWORD_RESET_EXPIRATION_MINUTES not specified")
	}
	passwordResetExpirationMinutesInt, err := strconv.Atoi(passwordResetExpirationMinutes)
	if err != nil {
		return nil, fmt.Errorf("PASSWORD_RESET_EXPIRATION_MINUTES not a number")
	}
	ttl := time.Duration(passwordResetExpirationMinutesInt) * time.Minute
	passwordLength := utils.GetEnv("PASSWORD_LENGTH", "8")
	length, err := strconv.Atoi(passwordLength)
	if err != nil {
		return nil, fmt.Errorf("PASSWORD_LENGTH not a number")
	}
	return &PasswordService{
		PasswordTokenTTL: &ttl,
		PasswordLength:   length,
	}, nil
}
