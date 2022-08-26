package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/pet_size"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

func SetPetSizeHandler(client *ent.Client, logger *zap.Logger, api *operations.BeAPI) {
	petSizeRepo := repositories.NewPetSizeRepository(client)
	petSizeHandler := NewPetSize(logger)

	api.PetSizeGetAllPetSizeHandler = petSizeHandler.GetAllPetSizeFunc(petSizeRepo)
	api.PetSizeEditPetSizeHandler = petSizeHandler.UpdatePetSizeByID(petSizeRepo)
	api.PetSizeDeletePetSizeHandler = petSizeHandler.DeletePetSizeByID(petSizeRepo)
	api.PetSizeCreateNewPetSizeHandler = petSizeHandler.CreatePetSizeFunc(petSizeRepo)
	api.PetSizeGetPetSizeHandler = petSizeHandler.GetPetSizeByID(petSizeRepo)
}

type PetSize struct {
	logger *zap.Logger
}

func NewPetSize(logger *zap.Logger) *PetSize {
	return &PetSize{
		logger: logger,
	}
}

func (ps PetSize) CreatePetSizeFunc(repository repositories.PetSizeRepository) pet_size.CreateNewPetSizeHandlerFunc {
	return func(p pet_size.CreateNewPetSizeParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		allPetSizes, err := repository.AllPetSizes(ctx)
		if err != nil {
			ps.logger.Error("Error while getting pet size", zap.Error(err))
			return pet_size.NewCreateNewPetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while creating pet size",
					},
				})
		}
		for _, petSize := range allPetSizes {
			if *p.NewPetSize.Name == petSize.Name {
				ps.logger.Error("Error while creating pet size", zap.Error(err))
				return pet_size.NewCreateNewPetSizeDefault(http.StatusInternalServerError).WithPayload(
					&models.Error{
						Data: &models.ErrorData{
							Message: "Error while creating pet size: the name already exist",
						},
					})
			}
		}
		petSize, err := repository.CreatePetSize(ctx, *p.NewPetSize)
		if err != nil {
			ps.logger.Error("Error while creating pet size", zap.Error(err))
			return pet_size.NewCreateNewPetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while creating pet size",
					},
				})
		}
		if petSize == nil {
			ps.logger.Error("Pet size is nil")
			return pet_size.NewCreateNewPetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while creating pet size",
					},
				})
		}
		id := int64(petSize.ID)
		return pet_size.NewCreateNewPetSizeCreated().WithPayload(&models.PetSizeResponse{
			ID:          &id,
			Name:        &petSize.Name,
			Size:        &petSize.Size,
			IsUniversal: false,
		},
		)
	}
}

func (ps PetSize) GetAllPetSizeFunc(repository repositories.PetSizeRepository) pet_size.GetAllPetSizeHandlerFunc {
	return func(p pet_size.GetAllPetSizeParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petSizes, err := repository.AllPetSizes(ctx)
		if err != nil {
			ps.logger.Error("Error while getting pet size", zap.Error(err))
			return pet_size.NewGetAllPetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while getting pet size",
					},
				})
		}
		if len(petSizes) == 0 {
			return pet_size.NewGetAllPetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "No pet size found",
					},
				})
		}
		listOfPetSize := models.ListOfPetSizes{}
		for _, v := range petSizes {
			listOfPetSize = append(listOfPetSize, &models.PetSize{
				ID:          int64(v.ID),
				Name:        &v.Name,
				Size:        &v.Size,
				IsUniversal: v.IsUniversal,
			})
		}
		return pet_size.NewGetAllPetSizeOK().WithPayload(listOfPetSize)
	}
}

func (ps PetSize) GetPetSizeByID(repo repositories.PetSizeRepository) pet_size.GetPetSizeHandlerFunc {
	return func(p pet_size.GetPetSizeParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petSize, err := repo.PetSizeByID(ctx, int(p.PetSizeID))
		if err != nil {
			ps.logger.Error("Error while getting pet size by id", zap.Error(err))
			return pet_size.NewGetPetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while getting pet size",
					},
				})
		}
		if petSize == nil {
			return pet_size.NewGetPetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while getting pet size",
					},
				})
		}
		id := int64(petSize.ID)
		return pet_size.NewGetPetSizeOK().WithPayload(&models.PetSizeResponse{ID: &id, Name: &petSize.Name, Size: &petSize.Size})
	}
}

func (ps PetSize) DeletePetSizeByID(repo repositories.PetSizeRepository) pet_size.DeletePetSizeHandlerFunc {
	return func(p pet_size.DeletePetSizeParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		err := repo.DeletePetSizeByID(ctx, int(p.PetSizeID))
		if err != nil {
			ps.logger.Error("Error while deleting pet size by id", zap.Error(err))
			return pet_size.NewDeletePetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while deleting pet size",
					},
				})
		}
		return pet_size.NewDeletePetSizeOK().WithPayload("Pet size deleted")
	}
}

func (ps PetSize) UpdatePetSizeByID(repo repositories.PetSizeRepository) pet_size.EditPetSizeHandlerFunc {
	return func(p pet_size.EditPetSizeParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petSize, err := repo.UpdatePetSizeByID(ctx, int(p.PetSizeID), p.EditPetSize)
		if err != nil {
			ps.logger.Error("Error while updating pet size by id", zap.Error(err))
			return pet_size.NewEditPetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while updating pet size",
					},
				})
		}
		if petSize == nil {
			ps.logger.Error("Error while updating pet size by id", zap.Error(err))
			return pet_size.NewEditPetSizeDefault(http.StatusInternalServerError).WithPayload(
				&models.Error{
					Data: &models.ErrorData{
						Message: "Error while updating pet size",
					},
				})
		}

		id := int64(petSize.ID)
		return pet_size.NewEditPetSizeOK().WithPayload(&models.PetSizeResponse{ID: &id, Name: &petSize.Name, Size: &petSize.Size})
	}
}
