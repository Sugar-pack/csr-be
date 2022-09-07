package repositories

import (
	"context"
	"errors"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type CategoryRepository interface {
	CreateCategory(ctx context.Context, newCategory models.CreateNewCategory) (*ent.Category, error)
	AllCategories(ctx context.Context, limit, offset int, orderBy, orderColumn string) ([]*ent.Category, error)
	AllCategoriesTotal(ctx context.Context) (int, error)
	CategoryByID(ctx context.Context, id int) (*ent.Category, error)
	DeleteCategoryByID(ctx context.Context, id int) error
	UpdateCategory(ctx context.Context, id int, update models.UpdateCategoryRequest) (*ent.Category, error)
}

var fieldsToOrderCategories = []string{
	category.FieldID,
	category.FieldName,
}

type categoryRepository struct {
	client *ent.Client
}

func NewCategoryRepository(client *ent.Client) CategoryRepository {
	return &categoryRepository{
		client: client,
	}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, newCategory models.CreateNewCategory) (*ent.Category, error) {
	return r.client.Category.Create().
		SetName(*newCategory.Name).
		SetMaxReservationUnits(*newCategory.MaxReservationUnits).
		SetMaxReservationTime(*newCategory.MaxReservationTime).
		SetHasSubcategory(*newCategory.HasSubcategory).
		Save(ctx)
}

func (r *categoryRepository) AllCategoriesTotal(ctx context.Context) (int, error) {
	return r.client.Category.Query().Count(ctx)
}

func (r *categoryRepository) AllCategories(ctx context.Context, limit, offset int,
	orderBy, orderColumn string) ([]*ent.Category, error) {
	if !utils.IsOrderField(orderColumn, fieldsToOrderCategories) {
		return nil, errors.New("wrong column to order by")
	}
	orderFunc, err := utils.GetOrderFunc(orderBy, orderColumn)
	if err != nil {
		return nil, err
	}
	return r.client.Category.Query().Order(orderFunc).Limit(limit).Offset(offset).All(ctx)
}

func (r *categoryRepository) CategoryByID(ctx context.Context, id int) (*ent.Category, error) {
	return r.client.Category.Query().Where(category.ID(id)).Only(ctx)
}

func (r *categoryRepository) DeleteCategoryByID(ctx context.Context, id int) error {
	return r.client.Category.DeleteOneID(id).Exec(ctx)
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, id int, update models.UpdateCategoryRequest) (*ent.Category, error) {
	categoryUpdate := r.client.Category.UpdateOneID(id)
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
