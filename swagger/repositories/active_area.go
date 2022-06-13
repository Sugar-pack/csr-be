package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

type ActiveAreaRepository interface {
	AllActiveAreas(ctx context.Context) ([]*ent.ActiveArea, error)
}

type activeAreaRepository struct {
	client *ent.Client
}

func NewActiveAreaRepository(client *ent.Client) ActiveAreaRepository {
	return &activeAreaRepository{client: client}
}

func (r *activeAreaRepository) AllActiveAreas(ctx context.Context) ([]*ent.ActiveArea, error) {
	return r.client.ActiveArea.Query().All(ctx)
}
