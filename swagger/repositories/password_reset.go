package repositories

import (
	"context"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/passwordreset"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type PasswordResetRepository interface {
	CreateToken(ctx context.Context, token string, ttl time.Time, userID int) error
	GetToken(ctx context.Context, token string) (*ent.PasswordReset, error)
	DeleteToken(ctx context.Context, token string) error
}

type passwordResetRepository struct {
}

func (p *passwordResetRepository) CreateToken(ctx context.Context, token string, ttl time.Time, userID int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	tokens, err := tx.PasswordReset.Query().QueryUsers().Where(user.IDEQ(userID)).QueryPasswordReset().All(ctx)
	if err != nil {
		return err
	}
	for _, t := range tokens {
		if errDelete := tx.PasswordReset.DeleteOneID(t.ID).Exec(ctx); errDelete != nil {
			return errDelete
		}
	}
	_, err = tx.PasswordReset.Create().SetToken(token).SetTTL(ttl).SetUsersID(userID).Save(ctx)
	return err
}

func (p *passwordResetRepository) GetToken(ctx context.Context, token string) (*ent.PasswordReset, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.PasswordReset.Query().Where(passwordreset.TokenEQ(token)).WithUsers().Only(ctx)
}

func (p *passwordResetRepository) DeleteToken(ctx context.Context, token string) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.PasswordReset.Delete().Where(passwordreset.TokenEQ(token)).Exec(ctx)
	return err
}

func NewPasswordResetRepository() PasswordResetRepository {
	return &passwordResetRepository{}
}
