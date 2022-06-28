package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/petkind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type PetKindRepository interface {
	CreatePetKind(ctx context.Context, ps models.PetKind) (*ent.PetKind, error)
	PetKindByID(ctx context.Context, id int) (*ent.PetKind, error)
	AllPetKinds(ctx context.Context) ([]*ent.PetKind, error)
	DeletePetKindByID(ctx context.Context, id int) error
	UpdatePetKindByID(ctx context.Context, id int, newPetKind *models.PetKind) (*ent.PetKind, error)
}
type petKindRepository struct {
	client *ent.Client
}

func NewPetKindRepository(client *ent.Client) PetKindRepository {
	return &petKindRepository{
		client: client,
	}
}
func (pkRepo petKindRepository) DeletePetKindByID(ctx context.Context, id int) error {
	return pkRepo.client.PetKind.DeleteOneID(id).Exec(ctx)
}

func (pkRepo petKindRepository) CreatePetKind(ctx context.Context, NewPetKind models.PetKind) (*ent.PetKind, error) {
	pk, err := pkRepo.client.PetKind.Create().
		SetName(*NewPetKind.Name).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	res, err := pkRepo.client.PetKind.Query().Where(petkind.ID(pk.ID)).WithEquipments().Only(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (pkRepo petKindRepository) PetKindByID(ctx context.Context, id int) (*ent.PetKind, error) {
	result, err := pkRepo.client.PetKind.Query().Where(petkind.ID(id)).WithEquipments().Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (pkRepo petKindRepository) AllPetKinds(ctx context.Context) ([]*ent.PetKind, error) {
	res, err := pkRepo.client.PetKind.Query().WithEquipments().All(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (pkRepo petKindRepository) UpdatePetKindByID(ctx context.Context, id int, newPetKind *models.PetKind) (*ent.PetKind, error) {
	oldPetSize, err := pkRepo.client.PetKind.Get(ctx, id)
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
