package email

type Sender interface {
	SendResetLink(email string, userName string, token string) error
	SendNewPassword(email string, userName string, password string) error
	SendRegistrationConfirmLink(email string, userName string, token string) error
}
