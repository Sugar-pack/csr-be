package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/petkind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type PetKindRepository interface {
	CreatePetKind(ctx context.Context, ps models.PetKind) (*ent.PetKind, error)
	PetKindByID(ctx context.Context, id int) (*ent.PetKind, error)
	AllPetKinds(ctx context.Context) ([]*ent.PetKind, error)
	DeletePetKindByID(ctx context.Context, id int) error
	UpdatePetKindByID(ctx context.Context, id int, newPetKind *models.PetKind) (*ent.PetKind, error)
}
type petKindRepository struct {
}

func NewPetKindRepository() PetKindRepository {
	return &petKindRepository{}
}
func (pkRepo petKindRepository) DeletePetKindByID(ctx context.Context, id int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	return tx.PetKind.DeleteOneID(id).Exec(ctx)
}

func (pkRepo petKindRepository) CreatePetKind(ctx context.Context, NewPetKind models.PetKind) (*ent.PetKind, error) {
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

func (pkRepo petKindRepository) PetKindByID(ctx context.Context, id int) (*ent.PetKind, error) {
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

func (pkRepo petKindRepository) AllPetKinds(ctx context.Context) ([]*ent.PetKind, error) {
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

func (pkRepo petKindRepository) UpdatePetKindByID(ctx context.Context, id int, newPetKind *models.PetKind) (*ent.PetKind, error) {
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
