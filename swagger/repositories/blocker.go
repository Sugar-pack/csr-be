package repositories

import (
	"context"
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

type BlockerRepository interface {
	SetIsBlockedUser(ctx context.Context, userId int, isBlocked bool) error
}

type blockerRepository struct {
	client *ent.Client
}

func NewBlockerRepository(client *ent.Client) BlockerRepository {
	return &blockerRepository{client: client}
}

func (r *blockerRepository) SetIsBlockedUser(ctx context.Context, userId int, isBlocked bool) error {
	user, err := r.client.User.Get(ctx, userId)
	if err != nil {
		return fmt.Errorf("status history error, failed to get user: %s", err)
	}
	_, err = r.client.User.UpdateOne(user).SetIsBlocked(isBlocked).Save(ctx)
	if err != nil {
		return fmt.Errorf("status history error, failed to  update user's isBlocked status: %s", err)
	}
	return nil
}
