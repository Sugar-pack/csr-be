package domain

import "context"

type PasswordResetService interface {
	SendResetPasswordLink(ctx context.Context, login string) error
	VerifyTokenAndSendPassword(ctx context.Context, tokenToVerify string) error
}

type RegistrationConfirmService interface {
	SendConfirmationLink(ctx context.Context, login string) error
	VerifyConfirmationToken(ctx context.Context, token string) error
	IsSendRequired() bool
}

type TokenManager interface {
	GenerateTokens(ctx context.Context, login, password string) (string, string, bool, error)
	RefreshToken(ctx context.Context, token string) (string, bool, error)
}
