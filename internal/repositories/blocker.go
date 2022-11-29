package repositories

import (
	"context"
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type blockerRepository struct {
}

func NewBlockerRepository() domain.BlockerRepository {
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
