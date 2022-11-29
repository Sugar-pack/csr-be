package email

import (
	"fmt"
	"net/smtp"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func NewWrapperSmtp(host, port, password string) *wrapperSmtp {
	return &wrapperSmtp{
		host:     host,
		port:     port,
		password: password,
	}
}

type wrapperSmtp struct {
	host, port string
	password   string
}

func (w *wrapperSmtp) Send(data *domain.SendData) error {
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := "From: " + fmt.Sprintf("\"%s\" <%s>", data.FromName, data.FromAddr) + "\n" +
		"To: " + data.ToAddr + "\n" +
		"Subject: " + data.Subject + "\n" +
		mime + "\n" +
		data.Text

	err := smtp.SendMail(fmt.Sprintf("%s:%s", w.host, w.port),
		smtp.PlainAuth("", data.FromAddr, w.password, w.host),
		data.FromAddr, []string{data.ToAddr}, []byte(msg))

	if err != nil {
		return fmt.Errorf("cant send email %w", err)
	}

	return nil
}
