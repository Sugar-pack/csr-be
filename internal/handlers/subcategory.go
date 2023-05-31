package handlers

import (
	"errors"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/subcategories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetSubcategoryHandler(logger *zap.Logger, api *operations.BeAPI) {
	subcategoryRepo := repositories.NewSubcategoryRepository()
	subcategoryHandler := NewSubcategory(logger)

	api.SubcategoriesCreateNewSubcategoryHandler = subcategoryHandler.CreateNewSubcategoryFunc(subcategoryRepo)
	api.SubcategoriesListSubcategoriesByCategoryIDHandler = subcategoryHandler.ListSubcategoriesFunc(subcategoryRepo)
	api.SubcategoriesGetSubcategoryByIDHandler = subcategoryHandler.GetSubcategoryByIDFunc(subcategoryRepo)
	api.SubcategoriesDeleteSubcategoryHandler = subcategoryHandler.DeleteSubcategoryFunc(subcategoryRepo)
	api.SubcategoriesUpdateSubcategoryHandler = subcategoryHandler.UpdateSubcategoryFunc(subcategoryRepo)
}

type Subcategory struct {
	logger *zap.Logger
}

func NewSubcategory(logger *zap.Logger) *Subcategory {
	return &Subcategory{
		logger: logger,
	}
}

func (s *Subcategory) CreateNewSubcategoryFunc(repository domain.SubcategoryRepository) subcategories.CreateNewSubcategoryHandlerFunc {
	return func(p subcategories.CreateNewSubcategoryParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		categoryID := int(p.CategoryID)
		createdSubcategory, err := repository.CreateSubcategory(ctx, categoryID, *p.NewSubcategory)
		if err != nil {
			s.logger.Error("failed to create new subcategory", zap.Error(err))
			return subcategories.NewCreateNewSubcategoryDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to create new subcategory"))
		}
		result, err := mapSubcategory(createdSubcategory)
		if err != nil {
			s.logger.Error("failed to map new subcategory", zap.Error(err))
			return subcategories.NewCreateNewSubcategoryDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to map new subcategory"))
		}
		return subcategories.NewCreateNewSubcategoryCreated().WithPayload(&models.SubcategoryResponse{
			Data: result,
		})
	}
}

func (s *Subcategory) ListSubcategoriesFunc(repository domain.SubcategoryRepository) subcategories.ListSubcategoriesByCategoryIDHandlerFunc {
	return func(p subcategories.ListSubcategoriesByCategoryIDParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		categoryID := int(p.CategoryID)
		subcategoriesList, err := repository.ListSubcategories(ctx, categoryID)
		if err != nil {
			s.logger.Error("failed to list subcategories by category id", zap.Error(err))
			return subcategories.NewListSubcategoriesByCategoryIDDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to list subcategories by category id"))
		}
		mappedSubcategories := make([]*models.Subcategory, len(subcategoriesList))
		for i, v := range subcategoriesList {
			mappedSubcategory, err := mapSubcategory(v)
			if err != nil {
				s.logger.Error("failed to map subcategories by category id", zap.Error(err))
				return subcategories.NewListSubcategoriesByCategoryIDDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("failed to map subcategories by category id"))
			}
			mappedSubcategories[i] = mappedSubcategory
		}
		return subcategories.NewListSubcategoriesByCategoryIDOK().WithPayload(mappedSubcategories)
	}
}

func (s *Subcategory) GetSubcategoryByIDFunc(repository domain.SubcategoryRepository) subcategories.GetSubcategoryByIDHandlerFunc {
	return func(p subcategories.GetSubcategoryByIDParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		subcategory, err := repository.SubcategoryByID(ctx, int(p.SubcategoryID))
		if err != nil {
			s.logger.Error("failed to get subcategory", zap.Error(err))
			return subcategories.NewGetSubcategoryByIDDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to get subcategory"))
		}
		result, err := mapSubcategory(subcategory)
		if err != nil {
			s.logger.Error("failed to map subcategory", zap.Error(err))
			return subcategories.NewGetSubcategoryByIDDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to map subcategory"))
		}
		return subcategories.NewGetSubcategoryByIDOK().
			WithPayload(&models.SubcategoryResponse{Data: result})
	}
}

func (s *Subcategory) DeleteSubcategoryFunc(repository domain.SubcategoryRepository) subcategories.DeleteSubcategoryHandlerFunc {
	return func(p subcategories.DeleteSubcategoryParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		err := repository.DeleteSubcategoryByID(ctx, int(p.SubcategoryID))
		if err != nil {
			s.logger.Error("failed to delete subcategory", zap.Error(err))
			return subcategories.NewDeleteSubcategoryDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to delete subcategory"))
		}
		return subcategories.NewDeleteSubcategoryOK().
			WithPayload("subcategory deleted")
	}
}

func (s *Subcategory) UpdateSubcategoryFunc(repository domain.SubcategoryRepository) subcategories.UpdateSubcategoryHandlerFunc {
	return func(p subcategories.UpdateSubcategoryParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		updateSubcategory, err := repository.UpdateSubcategory(ctx, int(p.SubcategoryID), *p.UpdateSubcategory)
		if err != nil {
			s.logger.Error("failed to update subcategory", zap.Error(err))
			return subcategories.NewUpdateSubcategoryDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to update subcategory"))
		}
		result, err := mapSubcategory(updateSubcategory)
		if err != nil {
			s.logger.Error("failed to map subcategory", zap.Error(err))
			return subcategories.NewUpdateSubcategoryDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("failed to map subcategory"))
		}
		return subcategories.NewUpdateSubcategoryOK().WithPayload(&models.SubcategoryResponse{
			Data: result,
		})
	}
}

func mapSubcategory(subcategory *ent.Subcategory) (*models.Subcategory, error) {
	if subcategory == nil {
		return nil, errors.New("subcategory is nil")
	}
	if subcategory.Edges.Category == nil {
		return nil, errors.New("category is nil")
	}
	subcategoryID := int64(subcategory.ID)
	categoryID := int64(subcategory.Edges.Category.ID)
	model := &models.Subcategory{
		Category:            &categoryID,
		ID:                  &subcategoryID,
		MaxReservationTime:  &subcategory.MaxReservationTime,
		MaxReservationUnits: &subcategory.MaxReservationUnits,
		Name:                &subcategory.Name,
	}
	return model, nil
}
