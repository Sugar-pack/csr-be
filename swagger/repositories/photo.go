package repositories

import (
	"context"
	"errors"
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/photo"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type PhotoRepository interface {
	CreatePhoto(ctx context.Context, p models.Photo) (*ent.Photo, error)
	PhotoByID(ctx context.Context, id string) (*ent.Photo, error)
	DeletePhotoByID(ctx context.Context, id string) error
}

type photoRepository struct {
}

func NewPhotoRepository() PhotoRepository {
	return &photoRepository{}
}

func (r *photoRepository) CreatePhoto(ctx context.Context, newPhoto models.Photo) (*ent.Photo, error) {
	if newPhoto.ID == nil || *newPhoto.ID == "" {
		return nil, errors.New("photo id must not be empty")
	}
	if newPhoto.FileName == "" {
		newPhoto.FileName = fmt.Sprintf("%s.jpg", *newPhoto.ID)
	}
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	p, err := tx.Photo.Create().
		SetID(*newPhoto.ID).
		SetFileName(newPhoto.FileName).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *photoRepository) PhotoByID(ctx context.Context, id string) (*ent.Photo, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := tx.Photo.Query().Where(photo.ID(id)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *photoRepository) DeletePhotoByID(ctx context.Context, id string) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Photo.Delete().Where(photo.ID(id)).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
