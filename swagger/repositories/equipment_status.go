package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type EquipmentStatusRepository interface {
	Create(ctx context.Context, name string) (*ent.Statuses, error)
	GetAll(ctx context.Context) ([]*ent.Statuses, error)
	Get(ctx context.Context, id int) (*ent.Statuses, error)
	Delete(ctx context.Context, id int) (*ent.Statuses, error)
}

type equipmentStatusRepository struct {
}

func NewEquipmentStatusRepository() EquipmentStatusRepository {
	return &equipmentStatusRepository{}
}

func (r *equipmentStatusRepository) Create(ctx context.Context, name string) (*ent.Statuses, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Statuses.Create().SetName(name).Save(ctx)
}

func (r *equipmentStatusRepository) GetAll(ctx context.Context) ([]*ent.Statuses, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Statuses.Query().All(ctx)
}

func (r *equipmentStatusRepository) Get(ctx context.Context, id int) (*ent.Statuses, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Statuses.Get(ctx, id)
}

func (r *equipmentStatusRepository) Delete(ctx context.Context, id int) (*ent.Statuses, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	statusToDelete, err := tx.Statuses.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	err = tx.Statuses.DeleteOne(statusToDelete).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return statusToDelete, nil
}
