package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"github.com/matcornic/hermes/v2"
)

func GenerateSendLinkReset(userName, websiteUrl, token string) (string, error) {
	return generateHtml(generateSendLinkReset(userName, websiteUrl, token))
}

func GenerateGetPasswordReset(userName, password string) (string, error) {
	return generateHtml(generateGetPasswordReset(userName, password))
}

type RegistrationConfirmData struct {
	Link string
}

//go:embed templates/registration-confirm/index.html
var content embed.FS

func GenerateRegistrationConfirmMessage(_, websiteUrl, token, htmlTemplatePath string) (string, error) {
	tmpl, err := template.ParseFS(content, htmlTemplatePath)
	if err != nil {
		return "", err
	}

	// Create the registration confirmation URL
	confirmationUrl := fmt.Sprintf("%sapi/registration_confirm/%s", websiteUrl, token)

	// Prepare the data to insert into the template
	data := RegistrationConfirmData{
		Link: confirmationUrl,
	}

	// Create a buffer to store the output of the template execution
	var tpl bytes.Buffer

	// Execute the template and write the output to the buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return "", err
	}

	// Convert the buffer contents to a string and return
	return tpl.String(), nil
}

func GenerateEmailConfirmMessage(userName, websiteUrl, token string) (string, error) {
	return generateHtml(generateEmailConfirmMessage(userName, websiteUrl, token))
}

func generateSendLinkReset(userName, websiteUrl, token string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: userName,
			Intros: []string{
				"Вы получили это электронное письмо, потому что был получен запрос на сброс пароля для учетной записи в сервисе Лёнькин Кот.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Нажмите кнопку ниже, чтобы сбросить свой пароль:",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Сбросить пароль",
						Link:  fmt.Sprintf("%sapi/password_reset/%s", websiteUrl, token),
					},
				},
			},
			Outros: []string{
				"Если вы не запрашивали сброс пароля, никаких дальнейших действий с вашей стороны не требуется.",
			},
			Signature: "Спасибо",
		},
	}
}

func generateGetPasswordReset(userName, password string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: userName,
			Intros: []string{
				"Вы получили это электронное письмо, потому что был получен запрос на сброс пароля для учетной записи в сервисе Лёнькин Кот.",
				"Ваш новый пароль:",
				password,
			},
			Outros: []string{
				"Если вы не запрашивали сброс пароля, никаких дальнейших действий с вашей стороны не требуется.",
			},
			Signature: "Спасибо",
		},
	}
}

func generateRegistrationConfirmMessage(userName, websiteUrl, token string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: userName,
			Intros: []string{
				"Вы получили это электронное письмо, потому что зарегистрировали учетную запись в сервисе Лёнькин Кот.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Нажмите кнопку ниже для подтверждения регистрации:",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Подтвердить",
						Link:  fmt.Sprintf("%sapi/registration_confirm/%s", websiteUrl, token),
					},
				},
			},
			Outros: []string{
				"Если вы не регистрировались, никаких дальнейших действий с вашей стороны не требуется.",
			},
			Signature: "Спасибо",
		},
	}
}

func generateEmailConfirmMessage(userName, websiteUrl, token string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name: userName,
			Intros: []string{
				"Вы получили это электронное письмо, потому что изменили адрес электронной почты в сервисе Лёнькин Кот.",
			},
			Actions: []hermes.Action{
				{
					Instructions: "Нажмите кнопку ниже для подтверждения нового адреса электронной почты:",
					Button: hermes.Button{
						Color: "#DC4D2F",
						Text:  "Подтвердить",
						Link:  fmt.Sprintf("%sapi/v1/email_confirm/%s", websiteUrl, token),
					},
				},
			},
			Outros: []string{
				"Если вы не изменяли адрес электронной почты, никаких дальнейших действий с вашей стороны не требуется.",
			},
			Signature: "Спасибо",
		},
	}
}

func generateHtml(he hermes.Email) (string, error) {
	year := time.Now().Year()
	h := hermes.Hermes{
		Product: hermes.Product{
			Name:      "Lyonkin Kot",
			Logo:      "http://lyonkinkot.ru/template/images/logo.png",
			Copyright: fmt.Sprintf("Copyright © %d Lyonkin Kot. All rights reserved.", year),
		},
	}

	return h.GenerateHTML(he)
}
