package email

import (
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

const (
	defaultRegistrationConfirmPath = "templates/registration-confirm/index.html"
)

type sender struct {
	websiteUrl              string
	senderName              string
	senderEmail             string
	isRequiredToSend        bool
	client                  domain.SMTPClient
	RegistrationConfirmPath string
}

func NewSenderSmtp(config config.Email, client domain.SMTPClient) domain.Sender {
	return &sender{
		websiteUrl:              config.SenderWebsiteUrl,
		senderName:              config.SenderFromName,
		senderEmail:             config.SenderFromAddress,
		isRequiredToSend:        config.IsSendRequired,
		client:                  client,
		RegistrationConfirmPath: defaultRegistrationConfirmPath,
	}
}

func (c *sender) IsSendRequired() bool {
	return c.isRequiredToSend
}

func (c *sender) SendResetLink(email string, userName string, token string) error {
	if c.isRequiredToSend == false {
		return nil
	}

	text, err := GenerateSendLinkReset(userName, c.websiteUrl, token)
	if err != nil {
		return fmt.Errorf("cant generate email %w", err)
	}

	sendData := &domain.SendData{
		FromName: c.senderName,
		FromAddr: c.senderEmail,
		Subject:  "Password Reset",
		ToAddr:   email,
		Text:     text,
	}

	err = c.client.Send(sendData)
	if err != nil {
		return fmt.Errorf("cant send email %w", err)
	}

	return err
}

func (c *sender) SendNewPassword(email string, userName string, password string) error {
	if c.isRequiredToSend == false {
		return nil
	}

	text, err := GenerateGetPasswordReset(userName, password)
	if err != nil {
		return fmt.Errorf("cant generate email %w", err)
	}
	sendData := &domain.SendData{
		FromName: c.senderName,
		FromAddr: c.senderEmail,
		Subject:  "New Password",
		ToAddr:   email,
		Text:     text,
	}
	err = c.client.Send(sendData)
	if err != nil {
		return fmt.Errorf("cant send email %w", err)
	}

	return nil
}

func (c *sender) SendRegistrationConfirmLink(email string, userName string, token string) error {
	if c.isRequiredToSend == false {
		return nil
	}
	text, err := GenerateRegistrationConfirmMessage(userName, c.websiteUrl, token, c.RegistrationConfirmPath)
	if err != nil {
		return fmt.Errorf("cant generate email %w", err)
	}
	sendData := &domain.SendData{
		FromName: c.senderName,
		FromAddr: c.senderEmail,
		Subject:  "Registration confirmation",
		ToAddr:   email,
		Text:     text,
	}
	err = c.client.Send(sendData)
	if err != nil {
		return fmt.Errorf("cant send email %w", err)
	}

	return err
}

func (c *sender) SendEmailConfirmationLink(email string, userName string, token string) error {
	if !c.isRequiredToSend {
		return nil
	}

	text, err := GenerateEmailConfirmMessage(userName, c.websiteUrl, token)
	if err != nil {
		return fmt.Errorf("cant generate message for new email confirmation  %w", err)
	}
	sendData := &domain.SendData{
		FromName: c.senderName,
		FromAddr: c.senderEmail,
		Subject:  "Email confirmation",
		ToAddr:   email,
		Text:     text,
	}
	err = c.client.Send(sendData)
	if err != nil {
		return fmt.Errorf("cant send email %w", err)
	}

	return err
}
