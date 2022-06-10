package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/petsize"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type PetSizeRepository interface {
	CreatePetSize(ctx context.Context, ps models.PetSize) (*ent.PetSize, error)
	PetSizeByID(ctx context.Context, id int) (*ent.PetSize, error)
	AllPetSizes(ctx context.Context) ([]*ent.PetSize, error)
	DeletePetSizeByID(ctx context.Context, id int) error
	UpdatePetSizeByID(ctx context.Context, id int, newPetSize *models.PetSize) (*ent.PetSize, error)
}
type petSizeRepository struct {
	client *ent.Client
}

func NewPetSizeRepository(client *ent.Client) PetSizeRepository {
	return &petSizeRepository{
		client: client,
	}
}
func (psRepo petSizeRepository) DeletePetSizeByID(ctx context.Context, id int) error {
	err := psRepo.client.PetSize.DeleteOneID(id).Exec(ctx)
	return err
}

func (psRepo petSizeRepository) CreatePetSize(ctx context.Context, NewPetSize models.PetSize) (*ent.PetSize, error) {
	ps, err := psRepo.client.PetSize.Create().
		SetName(*NewPetSize.Name).
		SetSize(*NewPetSize.Size).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	res, err := psRepo.client.PetSize.Query().Where(petsize.ID(ps.ID)).WithEquipments().Only(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (psRepo petSizeRepository) PetSizeByID(ctx context.Context, id int) (*ent.PetSize, error) {
	result, err := psRepo.client.PetSize.Query().Where(petsize.ID(id)).WithEquipments().Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (psRepo petSizeRepository) AllPetSizes(ctx context.Context) ([]*ent.PetSize, error) {
	res, err := psRepo.client.PetSize.Query().WithEquipments().All(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (psRepo petSizeRepository) UpdatePetSizeByID(ctx context.Context, id int, newPetSize *models.PetSize) (*ent.PetSize, error) {
	oldPetSize, err := psRepo.client.PetSize.Get(ctx, id)
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
