package handlers

import (
	"math"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/categories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetCategoryHandler(logger *zap.Logger, api *operations.BeAPI) {
	categoryRepo := repositories.NewCategoryRepository()
	categoryHandler := NewCategory(logger)

	api.CategoriesCreateNewCategoryHandler = categoryHandler.CreateNewCategoryFunc(categoryRepo)
	api.CategoriesGetCategoryByIDHandler = categoryHandler.GetCategoryByIDFunc(categoryRepo)
	api.CategoriesDeleteCategoryHandler = categoryHandler.DeleteCategoryFunc(categoryRepo)
	api.CategoriesGetAllCategoriesHandler = categoryHandler.GetAllCategoriesFunc(categoryRepo)
	api.CategoriesUpdateCategoryHandler = categoryHandler.UpdateCategoryFunc(categoryRepo)
}

type Category struct {
	logger *zap.Logger
}

func NewCategory(logger *zap.Logger) *Category {
	return &Category{
		logger: logger,
	}
}

func (c *Category) CreateNewCategoryFunc(repository domain.CategoryRepository) categories.CreateNewCategoryHandlerFunc {
	return func(s categories.CreateNewCategoryParams, _ *models.Principal) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		createdCategory, err := repository.CreateCategory(ctx, *s.NewCategory)
		if err != nil {
			c.logger.Error("cant create new category", zap.Error(err))
			return categories.NewCreateNewCategoryDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("cant create new category"))
		}
		return categories.NewCreateNewCategoryCreated().WithPayload(&models.CreateNewCategoryResponse{
			Data: mapCategory(createdCategory),
		})
	}
}

func (c *Category) GetAllCategoriesFunc(repository domain.CategoryRepository) categories.GetAllCategoriesHandlerFunc {
	return func(s categories.GetAllCategoriesParams, _ *models.Principal) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		limit := utils.GetValueByPointerOrDefaultValue(s.Limit, math.MaxInt)
		offset := utils.GetValueByPointerOrDefaultValue(s.Offset, 0)
		orderBy := utils.GetValueByPointerOrDefaultValue(s.OrderBy, utils.AscOrder)
		orderColumn := utils.GetValueByPointerOrDefaultValue(s.OrderColumn, category.FieldID)

		total, err := repository.AllCategoriesTotal(ctx)
		if err != nil {
			c.logger.Error("query total categories error", zap.Error(err))
			return categories.NewGetAllCategoriesDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("cant get total amount of categories"))
		}
		var allCategories []*ent.Category
		if total > 0 {
			filter := domain.CategoryFilter{
				Filter: domain.Filter{
					Limit:       int(limit),
					Offset:      int(offset),
					OrderBy:     orderBy,
					OrderColumn: orderColumn,
				},
			}
			if s.HasEquipments != nil {
				filter.HasEquipments = *s.HasEquipments
			}
			allCategories, err = repository.AllCategories(ctx, filter)
			if err != nil {
				c.logger.Error("query all category error", zap.Error(err))
				return categories.NewGetAllCategoriesDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("cant get all categories"))
			}
		}
		mappedCategories := make([]*models.Category, len(allCategories))
		for i, v := range allCategories {
			mappedCategories[i] = mapCategory(v)
		}
		totalCategories := int64(total)
		listOfCategories := &models.ListOfCategories{
			Items: mappedCategories,
			Total: &totalCategories,
		}
		return categories.NewGetAllCategoriesOK().WithPayload(listOfCategories)
	}
}

func (c *Category) GetCategoryByIDFunc(repository domain.CategoryRepository) categories.GetCategoryByIDHandlerFunc {
	return func(s categories.GetCategoryByIDParams, _ *models.Principal) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		category, err := repository.CategoryByID(ctx, int(s.CategoryID))
		if err != nil {
			c.logger.Error("failed to get category", zap.Error(err))
			return categories.NewGetCategoryByIDDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to get category"))
		}
		return categories.NewGetCategoryByIDOK().WithPayload(&models.GetCategoryByIDResponse{
			Data: mapCategory(category),
		})
	}
}

func (c *Category) DeleteCategoryFunc(repository domain.CategoryRepository) categories.DeleteCategoryHandlerFunc {
	return func(s categories.DeleteCategoryParams, _ *models.Principal) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		err := repository.DeleteCategoryByID(ctx, int(s.CategoryID))
		if err != nil {
			c.logger.Error("delete category failed", zap.Error(err))
			return categories.NewDeleteCategoryDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("delete category failed"))
		}
		return categories.NewDeleteCategoryOK().WithPayload("category deleted")
	}
}

func (c *Category) UpdateCategoryFunc(repository domain.CategoryRepository) categories.UpdateCategoryHandlerFunc {
	return func(s categories.UpdateCategoryParams, _ *models.Principal) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		updatedCategory, err := repository.UpdateCategory(ctx, int(s.CategoryID), *s.UpdateCategory)
		if err != nil {
			c.logger.Error("cant update category", zap.Error(err))
			return categories.NewUpdateCategoryDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("cant update category"))
		}

		return categories.NewUpdateCategoryOK().WithPayload(&models.UpdateCategoryResponse{
			Data: mapCategory(updatedCategory),
		})
	}
}

func mapCategory(category *ent.Category) *models.Category {
	if category == nil {
		return nil
	}
	id := int64(category.ID)
	return &models.Category{
		ID:                  &id,
		Name:                &category.Name,
		MaxReservationTime:  &category.MaxReservationTime,
		MaxReservationUnits: &category.MaxReservationUnits,
		HasSubcategory:      &category.HasSubcategory,
	}
}
