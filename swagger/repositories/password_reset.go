package repositories

import (
	"context"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/passwordreset"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/user"
)

type PasswordResetRepository interface {
	CreateToken(ctx context.Context, token string, ttl time.Time, userID int) error
	GetToken(ctx context.Context, token string) (*ent.PasswordReset, error)
	DeleteToken(ctx context.Context, token string) error
}

type passwordResetRepository struct {
	client *ent.Client
}

func (p *passwordResetRepository) CreateToken(ctx context.Context, token string, ttl time.Time, userID int) error {
	tokens, err := p.client.PasswordReset.Query().QueryUsers().Where(user.IDEQ(userID)).QueryPasswordReset().All(ctx)
	if err != nil {
		return err
	}
	for _, t := range tokens {
		if errDelete := p.client.PasswordReset.DeleteOneID(t.ID).Exec(ctx); errDelete != nil {
			return errDelete
		}
	}
	_, err = p.client.PasswordReset.Create().SetToken(token).SetTTL(ttl).SetUsersID(userID).Save(ctx)
	return err
}

func (p *passwordResetRepository) GetToken(ctx context.Context, token string) (*ent.PasswordReset, error) {
	return p.client.PasswordReset.Query().Where(passwordreset.TokenEQ(token)).WithUsers().Only(ctx)
}

func (p *passwordResetRepository) DeleteToken(ctx context.Context, token string) error {
	_, err := p.client.PasswordReset.Delete().Where(passwordreset.TokenEQ(token)).Exec(ctx)
	return err
}

func NewPasswordResetRepository(client *ent.Client) PasswordResetRepository {
	return &passwordResetRepository{
		client: client,
	}
}
