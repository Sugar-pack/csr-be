package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/subcategory"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type SubcategoryRepository interface {
	CreateSubcategory(ctx context.Context, categoryID int, newSubcategory models.NewSubcategory) (*ent.Subcategory, error)
	ListSubcategories(ctx context.Context, categoryID int) ([]*ent.Subcategory, error)
	SubcategoryByID(ctx context.Context, id int) (*ent.Subcategory, error)
	DeleteSubcategoryByID(ctx context.Context, id int) error
	UpdateSubcategory(ctx context.Context, id int, update models.NewSubcategory) (*ent.Subcategory, error)
}

type subcategoryRepository struct {
	client *ent.Client
}

func NewSubcategoryRepository(client *ent.Client) SubcategoryRepository {
	return &subcategoryRepository{
		client: client,
	}
}

func (r *subcategoryRepository) CreateSubcategory(
	ctx context.Context, categoryID int, newSubcategory models.NewSubcategory) (*ent.Subcategory, error) {
	eqCategory, err := r.client.Category.Get(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	saved, err := r.client.Subcategory.Create().
		SetName(*newSubcategory.Name).
		SetCategory(eqCategory).
		Save(ctx)
	return r.client.Subcategory.Query().Where(subcategory.ID(saved.ID)).
		WithCategory().Only(ctx)
}

func (r *subcategoryRepository) ListSubcategories(ctx context.Context, categoryID int) ([]*ent.Subcategory, error) {
	return r.client.Subcategory.Query().
		QueryCategory().Where(category.IDEQ(categoryID)).
		QuerySubcategories().WithCategory().All(ctx)
}

func (r *subcategoryRepository) SubcategoryByID(ctx context.Context, id int) (*ent.Subcategory, error) {
	return r.client.Subcategory.Query().Where(subcategory.ID(id)).WithCategory().Only(ctx)
}

func (r *subcategoryRepository) DeleteSubcategoryByID(ctx context.Context, id int) error {
	return r.client.Subcategory.DeleteOneID(id).Exec(ctx)
}

func (r *subcategoryRepository) UpdateSubcategory(ctx context.Context, id int, update models.NewSubcategory) (*ent.Subcategory, error) {
	subcategoryToUpdate := r.client.Subcategory.UpdateOneID(id)
	if update.Name != nil {
		subcategoryToUpdate.SetName(*update.Name)
	}
	_, err := subcategoryToUpdate.Save(ctx)
	if err != nil {
		return nil, err
	}
	return r.client.Subcategory.Query().Where(subcategory.ID(id)).WithCategory().Only(ctx)
}
