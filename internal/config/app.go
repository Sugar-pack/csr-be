package config

import (
	"fmt"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/email"
)

type AppConfig struct {
	PasswordConfig *PasswordService
	PhotoService   *PhotoService
	JWTSecret      string
	EmailService   email.ServiceConfig
}

func SetupAppConfig() (*AppConfig, error) {
	jwtSecretKey := utils.GetEnv("JWT_SECRET_KEY", "")
	if jwtSecretKey == "" {
		return nil, fmt.Errorf("JWT_SECRET_KEY not specified")

	}

	emailServiceConfig, err := email.SetupEmailServiceConfig()
	if err != nil {
		return nil, err
	}

	passwordServiceConfig, err := NewPasswordService()
	if err != nil {
		return nil, err
	}

	photoServiceConfig := NewPhotoServiceConfig()

	return &AppConfig{
		PasswordConfig: passwordServiceConfig,
		PhotoService:   photoServiceConfig,
		JWTSecret:      jwtSecretKey,
		EmailService:   emailServiceConfig,
	}, nil
}
