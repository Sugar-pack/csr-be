package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/subcategory"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type subcategoryRepository struct {
}

func NewSubcategoryRepository() domain.SubcategoryRepository {
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
		SetMaxReservationUnits(*newSubcategory.MaxReservationUnits).
		SetMaxReservationTime(*newSubcategory.MaxReservationTime).
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
	if update.MaxReservationTime != nil {
		subcategoryToUpdate.SetMaxReservationTime(*update.MaxReservationTime)
	}
	if update.MaxReservationUnits != nil {
		subcategoryToUpdate.SetMaxReservationUnits(*update.MaxReservationUnits)
	}
	_, err = subcategoryToUpdate.Save(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Subcategory.Query().Where(subcategory.ID(id)).WithCategory().Only(ctx)
}

func (r *subcategoryRepository) SubcategoryByEquipmentID(ctx context.Context, equipmentID int) (*ent.Subcategory, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Subcategory.Query().
		QueryEquipments().Where(equipment.IDEQ(equipmentID)).QuerySubcategory().
		Only(ctx)
}
