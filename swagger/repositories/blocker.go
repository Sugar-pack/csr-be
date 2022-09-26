package repositories

import (
	"context"
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type BlockerRepository interface {
	SetIsBlockedUser(ctx context.Context, userId int, isBlocked bool) error
}

type blockerRepository struct {
}

func NewBlockerRepository() BlockerRepository {
	return &blockerRepository{}
}

func (r *blockerRepository) SetIsBlockedUser(ctx context.Context, userId int, isBlocked bool) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	user, err := tx.User.Get(ctx, userId)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	_, err = tx.User.UpdateOne(user).SetIsBlocked(isBlocked).Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to  update user's isBlocked status: %w", err)
	}
	return nil
}
