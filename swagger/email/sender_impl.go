package email

import (
	"fmt"
)

type senderImpl struct {
	websiteUrl  string
	senderName  string
	senderEmail string
	client      *wrapperSmtp
}

func NewSenderSmtp(websiteUrl, smtpHost, smtpPort, smtpPassword, senderEmail, senderName string) Sender {

	return &senderImpl{
		websiteUrl:  websiteUrl,
		senderName:  senderName,
		senderEmail: senderEmail,
		client:      NewWrapperSmtp(smtpHost, smtpPort, smtpPassword),
	}
}

func (c *senderImpl) SendResetLink(email string, userName string, token string) error {
	text, err := GenerateSendLinkReset(userName, c.websiteUrl, token)
	if err != nil {
		return fmt.Errorf("cant generate email %w", err)
	}
	sendData := &SendData{
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

func (c *senderImpl) SendNewPassword(email string, userName string, password string) error {
	text, err := GenerateGetPasswordReset(userName, password)
	if err != nil {
		return fmt.Errorf("cant generate email %w", err)
	}
	sendData := &SendData{
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

func (c *senderImpl) SendRegistrationConfirmLink(email string, userName string, token string) error {
	text, err := GenerateRegistrationConfirmMessage(userName, c.websiteUrl, token)
	if err != nil {
		return fmt.Errorf("cant generate email %w", err)
	}
	sendData := &SendData{
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
