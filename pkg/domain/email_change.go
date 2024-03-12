package domain

import "context"

type ChangeEmailService interface {
	SendEmailConfirmationLink(ctx context.Context, login, email string) error
	VerifyTokenAndChangeEmail(ctx context.Context, tokenToVerify string) error
}
