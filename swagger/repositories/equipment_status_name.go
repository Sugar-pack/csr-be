package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type EquipmentStatusNameRepository interface {
	Create(ctx context.Context, name string) (*ent.EquipmentStatusName, error)
	GetAll(ctx context.Context) ([]*ent.EquipmentStatusName, error)
	Get(ctx context.Context, id int) (*ent.EquipmentStatusName, error)
	Delete(ctx context.Context, id int) (*ent.EquipmentStatusName, error)
}

type equipmentStatusNameRepository struct{}

func NewEquipmentStatusNameRepository() EquipmentStatusNameRepository {
	return &equipmentStatusNameRepository{}
}

func (r *equipmentStatusNameRepository) Create(ctx context.Context, name string) (*ent.EquipmentStatusName, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.EquipmentStatusName.Create().SetName(name).Save(ctx)
}

func (r *equipmentStatusNameRepository) GetAll(ctx context.Context) ([]*ent.EquipmentStatusName, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.EquipmentStatusName.Query().All(ctx)
}

func (r *equipmentStatusNameRepository) Get(ctx context.Context, id int) (*ent.EquipmentStatusName, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.EquipmentStatusName.Get(ctx, id)
}

func (r *equipmentStatusNameRepository) Delete(ctx context.Context, id int) (*ent.EquipmentStatusName, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	statusToDelete, err := tx.EquipmentStatusName.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	err = tx.EquipmentStatusName.DeleteOne(statusToDelete).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return statusToDelete, nil
}
