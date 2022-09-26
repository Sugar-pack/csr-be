package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/subcategory"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type SubcategoryRepository interface {
	CreateSubcategory(ctx context.Context, categoryID int, newSubcategory models.NewSubcategory) (*ent.Subcategory, error)
	ListSubcategories(ctx context.Context, categoryID int) ([]*ent.Subcategory, error)
	SubcategoryByID(ctx context.Context, id int) (*ent.Subcategory, error)
	DeleteSubcategoryByID(ctx context.Context, id int) error
	UpdateSubcategory(ctx context.Context, id int, update models.NewSubcategory) (*ent.Subcategory, error)
}

type subcategoryRepository struct {
}

func NewSubcategoryRepository() SubcategoryRepository {
	return &subcategoryRepository{}
}

func (r *subcategoryRepository) CreateSubcategory(
	ctx context.Context, categoryID int, newSubcategory models.NewSubcategory) (*ent.Subcategory, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	eqCategory, err := tx.Category.Get(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	saved, err := tx.Subcategory.Create().
		SetName(*newSubcategory.Name).
		SetCategory(eqCategory).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Subcategory.Query().Where(subcategory.ID(saved.ID)).
		WithCategory().Only(ctx)
}

func (r *subcategoryRepository) ListSubcategories(ctx context.Context, categoryID int) ([]*ent.Subcategory, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Subcategory.Query().
		QueryCategory().Where(category.IDEQ(categoryID)).
		QuerySubcategories().WithCategory().All(ctx)
}

func (r *subcategoryRepository) SubcategoryByID(ctx context.Context, id int) (*ent.Subcategory, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Subcategory.Query().Where(subcategory.ID(id)).WithCategory().Only(ctx)
}

func (r *subcategoryRepository) DeleteSubcategoryByID(ctx context.Context, id int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	return tx.Subcategory.DeleteOneID(id).Exec(ctx)
}

func (r *subcategoryRepository) UpdateSubcategory(ctx context.Context, id int, update models.NewSubcategory) (*ent.Subcategory, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	subcategoryToUpdate := tx.Subcategory.UpdateOneID(id)
	if update.Name != nil {
		subcategoryToUpdate.SetName(*update.Name)
	}
	_, err = subcategoryToUpdate.Save(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Subcategory.Query().Where(subcategory.ID(id)).WithCategory().Only(ctx)
}
