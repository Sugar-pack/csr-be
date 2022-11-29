package domain

type PasswordGenerator interface {
	NewPassword() (string, error)
}
