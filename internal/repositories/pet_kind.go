package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/petkind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type petKindRepository struct {
}

func NewPetKindRepository() domain.PetKindRepository {
	return &petKindRepository{}
}
func (pkRepo petKindRepository) Delete(ctx context.Context, id int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	return tx.PetKind.DeleteOneID(id).Exec(ctx)
}

func (pkRepo petKindRepository) Create(ctx context.Context, NewPetKind models.PetKind) (*ent.PetKind, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	pk, err := tx.PetKind.Create().
		SetName(*NewPetKind.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	res, err := tx.PetKind.Query().Where(petkind.ID(pk.ID)).WithEquipments().Only(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (pkRepo petKindRepository) GetByID(ctx context.Context, id int) (*ent.PetKind, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := tx.PetKind.Query().Where(petkind.ID(id)).WithEquipments().Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (pkRepo petKindRepository) GetAll(ctx context.Context) ([]*ent.PetKind, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	res, err := tx.PetKind.Query().WithEquipments().All(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (pkRepo petKindRepository) Update(ctx context.Context, id int, newPetKind *models.PetKind) (*ent.PetKind, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	oldPetSize, err := tx.PetKind.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	edit := oldPetSize.Update()
	if *newPetKind.Name != "" {
		edit.SetName(*newPetKind.Name)
	}
	res, err := edit.Save(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}
