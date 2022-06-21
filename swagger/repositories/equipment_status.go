package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

type EquipmentStatusRepository interface {
	Create(ctx context.Context, name string) (*ent.Statuses, error)
	GetAll(ctx context.Context) ([]*ent.Statuses, error)
	Get(ctx context.Context, id int) (*ent.Statuses, error)
	Delete(ctx context.Context, id int) (*ent.Statuses, error)
}

type equipmentStatusRepository struct {
	client *ent.Client
}

func NewEquipmentStatusRepository(client *ent.Client) EquipmentStatusRepository {
	return &equipmentStatusRepository{
		client: client,
	}
}

func (r *equipmentStatusRepository) Create(ctx context.Context, name string) (*ent.Statuses, error) {
	return r.client.Statuses.Create().SetName(name).Save(ctx)
}

func (r *equipmentStatusRepository) GetAll(ctx context.Context) ([]*ent.Statuses, error) {
	return r.client.Statuses.Query().All(ctx)
}

func (r *equipmentStatusRepository) Get(ctx context.Context, id int) (*ent.Statuses, error) {
	return r.client.Statuses.Get(ctx, id)
}

func (r *equipmentStatusRepository) Delete(ctx context.Context, id int) (*ent.Statuses, error) {
	statusToDelete, err := r.client.Statuses.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	err = r.client.Statuses.DeleteOne(statusToDelete).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return statusToDelete, nil
}
