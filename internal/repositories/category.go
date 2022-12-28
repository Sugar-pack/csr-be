package repositories

import (
	"context"
	"errors"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

var fieldsToOrderCategories = []string{
	category.FieldID,
	category.FieldName,
}

type categoryRepository struct {
}

func NewCategoryRepository() domain.CategoryRepository {
	return &categoryRepository{}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, newCategory models.CreateNewCategory) (*ent.Category, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Category.Create().
		SetName(*newCategory.Name).
		SetMaxReservationUnits(*newCategory.MaxReservationUnits).
		SetMaxReservationTime(*newCategory.MaxReservationTime).
		SetHasSubcategory(*newCategory.HasSubcategory).
		Save(ctx)
}

func (r *categoryRepository) AllCategoriesTotal(ctx context.Context) (int, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return tx.Category.Query().Count(ctx)
}

func (r *categoryRepository) AllCategories(ctx context.Context, filter domain.CategoryFilter) ([]*ent.Category, error) {
	if !utils.IsValueInList(filter.OrderColumn, fieldsToOrderCategories) {
		return nil, errors.New("wrong column to order by")
	}

	orderFunc, err := utils.GetOrderFunc(filter.OrderBy, filter.OrderColumn)
	if err != nil {
		return nil, err
	}

	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}

	q := tx.Category.Query()

	if filter.HasEquipments {
		q = q.QueryEquipments().QueryCategory()
	}

	return q.Order(orderFunc).
		Limit(filter.Limit).
		Offset(filter.Offset).
		All(ctx)
}

func (r *categoryRepository) CategoryByID(ctx context.Context, id int) (*ent.Category, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Category.Query().Where(category.ID(id)).Only(ctx)
}

func (r *categoryRepository) DeleteCategoryByID(ctx context.Context, id int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	return tx.Category.DeleteOneID(id).Exec(ctx)
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, id int, update models.UpdateCategoryRequest) (*ent.Category, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	categoryUpdate := tx.Category.UpdateOneID(id)
	if update.Name != nil {
		categoryUpdate.SetName(*update.Name)
	}
	if update.MaxReservationUnits != nil {
		categoryUpdate.SetMaxReservationUnits(*update.MaxReservationUnits)
	}
	if update.MaxReservationTime != nil {
		categoryUpdate.SetMaxReservationTime(*update.MaxReservationTime)
	}
	if update.HasSubcategory != nil {
		categoryUpdate.SetHasSubcategory(*update.HasSubcategory)
	}
	return categoryUpdate.Save(ctx)
}
