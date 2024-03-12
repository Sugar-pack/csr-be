package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/pet_kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetPetKindHandler(logger *zap.Logger, api *operations.BeAPI) {
	petKindRepo := repositories.NewPetKindRepository()
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

func (pk PetKind) CreatePetKindFunc(repository domain.PetKindRepository) pet_kind.CreateNewPetKindHandlerFunc {
	return func(p pet_kind.CreateNewPetKindParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petKind, err := repository.Create(ctx, *p.NewPetKind)
		if err != nil {
			pk.logger.Error(messages.ErrCreatePetKind, zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrCreatePetKind, err.Error()))
		}
		if petKind == nil {
			pk.logger.Error("Pet kind is nil")
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrCreatePetKind, ""))
		}
		id := int64(petKind.ID)
		return pet_kind.NewCreateNewPetKindCreated().WithPayload(&models.PetKindResponse{
			ID:   &id,
			Name: &petKind.Name,
		},
		)
	}
}

func (pk PetKind) GetAllPetKindFunc(repository domain.PetKindRepository) pet_kind.GetAllPetKindsHandlerFunc {
	return func(p pet_kind.GetAllPetKindsParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petKinds, err := repository.GetAll(ctx)
		if err != nil {
			pk.logger.Error(messages.ErrGetPetKind, zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrGetPetKind, err.Error()))
		}
		if len(petKinds) == 0 {
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrPetKindNotFound, ""))
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

func (pk PetKind) GetPetKindsByID(repo domain.PetKindRepository) pet_kind.GetPetKindHandlerFunc {
	return func(p pet_kind.GetPetKindParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petKind, err := repo.GetByID(ctx, int(p.PetKindID))
		if err != nil {
			pk.logger.Error(messages.ErrGetPetKind, zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrGetPetKind, err.Error()))
		}
		if petKind == nil {
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrGetPetKind, ""))
		}
		id := int64(petKind.ID)
		return pet_kind.NewGetPetKindOK().WithPayload(&models.PetKindResponse{ID: &id, Name: &petKind.Name})
	}
}

func (pk PetKind) DeletePetKindByID(repo domain.PetKindRepository) pet_kind.DeletePetKindHandlerFunc {
	return func(p pet_kind.DeletePetKindParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		err := repo.Delete(ctx, int(p.PetKindID))
		if err != nil {
			pk.logger.Error(messages.ErrDeletePetKind, zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrDeletePetKind, err.Error()))
		}
		return pet_kind.NewDeletePetKindOK().WithPayload(messages.MsgPetKindDeleted)
	}
}

func (pk PetKind) UpdatePetKindByID(repo domain.PetKindRepository) pet_kind.EditPetKindHandlerFunc {
	return func(p pet_kind.EditPetKindParams, _ *models.Principal) middleware.Responder {
		ctx := p.HTTPRequest.Context()
		petSize, err := repo.Update(ctx, int(p.PetKindID), p.EditPetKind)
		if err != nil {
			pk.logger.Error(messages.ErrUpdatePetKind, zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrUpdatePetKind, err.Error()))
		}
		if petSize == nil {
			pk.logger.Error(messages.ErrUpdatePetKind, zap.Error(err))
			return pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError).
				WithPayload(buildInternalErrorPayload(messages.ErrUpdatePetKind, ""))
		}

		id := int64(petSize.ID)
		return pet_kind.NewEditPetKindOK().WithPayload(&models.PetKindResponse{ID: &id, Name: &petSize.Name})
	}
}
