package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/pet_kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

func SetPetKindHandler(client *ent.Client, logger *zap.Logger, api *operations.BeAPI) {
	petKindRepo := repositories.NewPetKindRepository(client)
	petKindHandler := NewPetKind(logger)

	api.PetKindGetAllPetKindsHandler = petKindHandler.GetAllPetKindFunc(petKindRepo)
	api.PetKindEditPetKindHandler = petKindHandler.UpdatePetKindByID(petKindRepo)
	api.PetKindDeletePetKindHandler = petKindHandler.DeletePetKindByID(petKindRepo)
	api.PetKindCreateNewPetKindHandler = petKindHandler.CreatePetKindFunc(petKindRepo)
	api.PetKindGetPetKindHandler = petKindHandler.GetPetKindsByID(petKindRepo)

}

type PetKind struct {
	logger *zap.Logger
}

func NewPetKind(logger *zap.Logger) *PetKind {
	return &PetKind{
		logger: logger,
	}
}

func (pk PetKind) CreatePetKindFunc(repository repositories.PetKindRepository) pet_kind.CreateNewPetKindHandlerFunc {
	return func(p pet_kind.CreateNewPetKindParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petKind, err := repository.CreatePetKind(ctx, *p.NewPetKind)
		if err != nil {
			pk.logger.Error("Error while creating pet kind", zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while creating pet kind"))
		}
		if petKind == nil {
			pk.logger.Error("Pet kind is nil")
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while creating pet kind"))
		}
		id := int64(petKind.ID)
		return pet_kind.NewCreateNewPetKindCreated().WithPayload(&models.PetKindResponse{
			ID:   &id,
			Name: &petKind.Name,
		},
		)
	}
}

func (pk PetKind) GetAllPetKindFunc(repository repositories.PetKindRepository) pet_kind.GetAllPetKindsHandlerFunc {
	return func(p pet_kind.GetAllPetKindsParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petKinds, err := repository.AllPetKinds(ctx)
		if err != nil {
			pk.logger.Error("Error while getting pet kind", zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while getting pet kind"))
		}
		if len(petKinds) == 0 {
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("No pet kind found"))
		}
		listOfPetKind := models.ListOfPetKinds{}
		for _, v := range petKinds {
			id64 := int64(v.ID)
			pk := models.PetKindResponse{ID: &id64, Name: &v.Name}
			listOfPetKind = append(listOfPetKind, &pk)
		}
		return pet_kind.NewGetAllPetKindsOK().WithPayload(listOfPetKind)
	}
}

func (pk PetKind) GetPetKindsByID(repo repositories.PetKindRepository) pet_kind.GetPetKindHandlerFunc {
	return func(p pet_kind.GetPetKindParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petKind, err := repo.PetKindByID(ctx, int(p.PetKindID))
		if err != nil {
			pk.logger.Error("Error while getting pet kind by id", zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while getting pet kind"))
		}
		if petKind == nil {
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while getting pet kind"))
		}
		id := int64(petKind.ID)
		return pet_kind.NewGetPetKindOK().WithPayload(&models.PetKindResponse{ID: &id, Name: &petKind.Name})
	}
}

func (pk PetKind) DeletePetKindByID(repo repositories.PetKindRepository) pet_kind.DeletePetKindHandlerFunc {
	return func(p pet_kind.DeletePetKindParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		err := repo.DeletePetKindByID(ctx, int(p.PetKindID))
		if err != nil {
			pk.logger.Error("Error while deleting pet kind by id", zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while deleting pet kind"))
		}
		return pet_kind.NewDeletePetKindOK().WithPayload("Pet kind deleted")
	}
}

func (pk PetKind) UpdatePetKindByID(repo repositories.PetKindRepository) pet_kind.EditPetKindHandlerFunc {
	return func(p pet_kind.EditPetKindParams, access interface{}) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petSize, err := repo.UpdatePetKindByID(ctx, int(p.PetKindID), p.EditPetKind)
		if err != nil {
			pk.logger.Error("Error while updating pet kind by id", zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while updating pet kind"))
		}
		if petSize == nil {
			pk.logger.Error("Error while updating pet kind by id", zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while updating pet kind"))
		}

		id := int64(petSize.ID)
		return pet_kind.NewEditPetKindOK().WithPayload(&models.PetKindResponse{ID: &id, Name: &petSize.Name})
	}
}
