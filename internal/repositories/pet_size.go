package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/petsize"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type petSizeRepository struct {
}

func NewPetSizeRepository() domain.PetSizeRepository {
	return &petSizeRepository{}
}
func (psRepo petSizeRepository) Delete(ctx context.Context, id int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	return tx.PetSize.DeleteOneID(id).Exec(ctx)
}

func (psRepo petSizeRepository) Create(ctx context.Context, NewPetSize models.PetSize) (*ent.PetSize, error) {
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

func (psRepo petSizeRepository) GetByID(ctx context.Context, id int) (*ent.PetSize, error) {
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

func (psRepo petSizeRepository) GetAll(ctx context.Context) ([]*ent.PetSize, error) {
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

func (psRepo petSizeRepository) Update(ctx context.Context, id int, newPetSize *models.PetSize) (*ent.PetSize, error) {
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
