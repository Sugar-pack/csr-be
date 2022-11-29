package domain

type Sender interface {
	SendResetLink(email string, userName string, token string) error
	SendNewPassword(email string, userName string, password string) error
	SendRegistrationConfirmLink(email string, userName string, token string) error
	IsSendRequired() bool
}

type SMTPClient interface {
	Send(data *SendData) error
}

type SendData struct {
	FromName, FromAddr, ToAddr, Subject, Text string
}
