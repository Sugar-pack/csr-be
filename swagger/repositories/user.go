package repositories

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

type UserRepository interface {
	SetUserRole(ctx context.Context, userId int, roleId int) (*ent.User, error)
}

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) SetUserRole(ctx context.Context, userId int, roleId int) (foundUser *ent.User, resultError error) {
	tx, err := r.client.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	defer func(tx *ent.Tx) {
		err := tx.Commit()
		if err != nil {
			resultError = err
			foundUser = nil
		}
	}(tx)

	user, err := tx.User.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	role, err := tx.Role.Get(ctx, roleId)
	if err != nil {
		return nil, err
	}

	foundUser, err = r.client.User.UpdateOne(user).SetRole(role).Save(ctx)
	if err != nil {
		return nil, err
	}

	return foundUser, nil
}
