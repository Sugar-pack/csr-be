package repositories

import (
	"context"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/registrationconfirm"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/user"
)

type RegistrationConfirmRepository interface {
	CreateToken(ctx context.Context, token string, ttl time.Time, userID int) error
	GetToken(ctx context.Context, token string) (*ent.RegistrationConfirm, error)
	DeleteToken(ctx context.Context, token string) error
}

type registrationConfirmRepository struct {
	client *ent.Client
}

func (rc *registrationConfirmRepository) CreateToken(ctx context.Context, token string, ttl time.Time, userID int) error {
	tokens, err := rc.client.RegistrationConfirm.Query().QueryUsers().Where(user.IDEQ(userID)).QueryRegistrationConfirm().All(ctx)
	if err != nil {
		return err
	}
	for _, t := range tokens {
		if errDelete := rc.client.RegistrationConfirm.DeleteOneID(t.ID).Exec(ctx); errDelete != nil {
			return errDelete
		}
	}
	_, err = rc.client.RegistrationConfirm.Create().SetToken(token).SetTTL(ttl).SetUsersID(userID).Save(ctx)
	return err
}

func (rc *registrationConfirmRepository) GetToken(ctx context.Context, token string) (*ent.RegistrationConfirm, error) {
	return rc.client.RegistrationConfirm.Query().Where(registrationconfirm.TokenEQ(token)).WithUsers().Only(ctx)
}

func (rc *registrationConfirmRepository) DeleteToken(ctx context.Context, token string) error {
	_, err := rc.client.RegistrationConfirm.Delete().Where(registrationconfirm.TokenEQ(token)).Exec(ctx)
	return err
}

func NewRegistrationConfirmRepository(client *ent.Client) RegistrationConfirmRepository {
	return &registrationConfirmRepository{
		client: client,
	}
}
