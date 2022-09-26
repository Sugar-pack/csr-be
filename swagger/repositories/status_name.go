package repositories

import (
	"context"
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type StatusNameRepository interface {
	ListOfStatuses(ctx context.Context) ([]*ent.StatusName, error)
}

type statusNameRepository struct {
}

func NewStatusNameRepository() *statusNameRepository {
	return &statusNameRepository{}
}

func (r *statusNameRepository) ListOfStatuses(ctx context.Context) ([]*ent.StatusName, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	pointersStatuses, err := tx.StatusName.Query().Order(ent.Asc("id")).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get status names: %s", err)
	}
	return pointersStatuses, nil
}
