package email

import (
	"fmt"
	"os"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

type ServiceConfig struct {
	EmailServerHost   string
	ServerPort        string
	password          string
	SenderFromAddress string
	FromName          string
	SenderWebsiteUrl  string
}

func SetupEmailServiceConfig() (ServiceConfig, error) {
	emailSenderServerHost := os.Getenv("EMAIL_SENDER_SERVER_HOST")
	if emailSenderServerHost == "" {
		return ServiceConfig{}, fmt.Errorf("EMAIL_SENDER_SERVER_HOST is not set")
	}
	emailSenderServerPort := os.Getenv("EMAIL_SENDER_SERVER_PORT")
	if emailSenderServerPort == "" {
		return ServiceConfig{}, fmt.Errorf("EMAIL_SENDER_SERVER_PORT is not set")
	}
	emailSenderPassword := os.Getenv("EMAIL_SENDER_PASSWORD")
	if emailSenderPassword == "" {
		return ServiceConfig{}, fmt.Errorf("EMAIL_SENDER_PASSWORD is not set")
	}
	emailSenderFromAddress := os.Getenv("EMAIL_SENDER_FROM_ADDRESS")
	if emailSenderFromAddress == "" {
		return ServiceConfig{}, fmt.Errorf("EMAIL_SENDER_FROM_ADDRESS is not set")
	}
	emailSenderFromName := os.Getenv("EMAIL_SENDER_FROM_NAME")
	if emailSenderFromName == "" {
		return ServiceConfig{}, fmt.Errorf("EMAIL_SENDER_FROM_NAME is not set")
	}
	emailSenderWebsiteUrl := utils.GetEnv("EMAIL_SENDER_WEBSITE_URL", "https://csr.golangforall.com/")
	return ServiceConfig{
		EmailServerHost:   emailSenderServerHost,
		ServerPort:        emailSenderServerPort,
		password:          emailSenderPassword,
		SenderFromAddress: emailSenderFromAddress,
		FromName:          emailSenderFromName,
		SenderWebsiteUrl:  emailSenderWebsiteUrl,
	}, nil
}
