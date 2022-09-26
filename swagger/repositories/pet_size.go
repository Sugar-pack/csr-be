package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/petsize"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type PetSizeRepository interface {
	CreatePetSize(ctx context.Context, ps models.PetSize) (*ent.PetSize, error)
	PetSizeByID(ctx context.Context, id int) (*ent.PetSize, error)
	AllPetSizes(ctx context.Context) ([]*ent.PetSize, error)
	DeletePetSizeByID(ctx context.Context, id int) error
	UpdatePetSizeByID(ctx context.Context, id int, newPetSize *models.PetSize) (*ent.PetSize, error)
}
type petSizeRepository struct {
}

func NewPetSizeRepository() PetSizeRepository {
	return &petSizeRepository{}
}
func (psRepo petSizeRepository) DeletePetSizeByID(ctx context.Context, id int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	return tx.PetSize.DeleteOneID(id).Exec(ctx)
}

func (psRepo petSizeRepository) CreatePetSize(ctx context.Context, NewPetSize models.PetSize) (*ent.PetSize, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	ps, err := tx.PetSize.Create().
		SetName(*NewPetSize.Name).
		SetSize(*NewPetSize.Size).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	res, err := tx.PetSize.Query().Where(petsize.ID(ps.ID)).WithEquipments().Only(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (psRepo petSizeRepository) PetSizeByID(ctx context.Context, id int) (*ent.PetSize, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := tx.PetSize.Query().Where(petsize.ID(id)).WithEquipments().Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (psRepo petSizeRepository) AllPetSizes(ctx context.Context) ([]*ent.PetSize, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	res, err := tx.PetSize.Query().WithEquipments().All(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (psRepo petSizeRepository) UpdatePetSizeByID(ctx context.Context, id int, newPetSize *models.PetSize) (*ent.PetSize, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	oldPetSize, err := tx.PetSize.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	edit := oldPetSize.Update()
	if *newPetSize.Name != "" {
		edit.SetName(*newPetSize.Name)
	}
	if *newPetSize.Size != "" {
		edit.SetSize(*newPetSize.Size)
	}
	res, err := edit.Save(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}
