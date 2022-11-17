package config

import (
	"fmt"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/email"
)

type AppConfig struct {
	PasswordConfig        *PasswordService
	PhotoService          *PhotoService
	JWTSecret             string
	EmailService          email.ServiceConfig
	OverdueTimeCheckHours time.Duration
}

func SetupAppConfig() (*AppConfig, error) {
	jwtSecretKey := utils.GetEnv("JWT_SECRET_KEY", "")
	if jwtSecretKey == "" {
		return nil, fmt.Errorf("JWT_SECRET_KEY not specified")
	}
	overdueTimeCheck := utils.GetEnv("OVERDUE_TIME_CHECK_HOURS", "4h")
	if overdueTimeCheck == "" {
		return nil, fmt.Errorf("OVERDUE_TIME_CHECK_HOURS not specified")
	}
	overdueDurHours, err := time.ParseDuration(overdueTimeCheck)
	if err != nil {
		return nil, fmt.Errorf("OVERDUE_TIME_CHECK_HOURS must be in duration format, e.x. 4h")
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
		PasswordConfig:        passwordServiceConfig,
		PhotoService:          photoServiceConfig,
		JWTSecret:             jwtSecretKey,
		EmailService:          emailServiceConfig,
		OverdueTimeCheckHours: overdueDurHours,
	}, nil
}
