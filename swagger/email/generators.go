package email

import (
	"fmt"
	"github.com/matcornic/hermes/v2"
)

func GenerateSendLinkReset(userName, host, token string) (string, error) {
	return generateHtml(generateSendLinkReset(userName, host, token))
}

func GenerateGetPasswordReset(userName, password string) (string, error) {
	return generateHtml(generateGetPasswordReset(userName, password))
}

func generateSendLinkReset(userName, host, token string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: userName,
			Intros: []string{
				"You have received this email because a password reset request for Lyonkin Kot account was received.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Click the button below to reset your password:",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Reset your password",
						Link:  fmt.Sprintf("https://%s/api/password_reset/%s", host, token),
					},
				},
			},
			Outros: []string{
				"If you did not request a password reset, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}

func generateGetPasswordReset(userName, password string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: userName,
			Intros: []string{
				"You have received this email because a password reset request for Lyonkin Kot account was received.",
				"This is your new password:",
				password,
			},
			Outros: []string{
				"If you did not request a password reset, no further action is required on your part.",
			},
			Signature: "Thanks",
		},
	}
}

func generateHtml(he hermes.Email) (string, error) {
	h := hermes.Hermes{
		Product: hermes.Product{
			Name:      "Lyonkin Kot",
			Logo:      "http://lyonkinkot.ru/template/images/logo.png",
			Copyright: "Copyright © 2022 Lyonkin Kot. All rights reserved.",
		},
	}

	return h.GenerateHTML(he)
}
