package repositories

import (
	"context"
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

type StatusNameRepository interface {
	ListOfStatuses(ctx context.Context) ([]*ent.StatusName, error)
}

type statusNameRepository struct {
	client *ent.Client
}

func NewStatusNameRepository(client *ent.Client) *statusNameRepository {
	return &statusNameRepository{client: client}
}

func (r *statusNameRepository) ListOfStatuses(ctx context.Context) ([]*ent.StatusName, error) {
	pointersStatuses, err := r.client.StatusName.Query().Order(ent.Asc("id")).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get status names: %s", err)
	}
	return pointersStatuses, nil
}
