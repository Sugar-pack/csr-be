package handlers

import (
	"errors"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/services"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
)

type Equipment struct {
	logger *zap.Logger
}

func NewEquipment(logger *zap.Logger) *Equipment {
	return &Equipment{
		logger: logger,
	}
}

func (c Equipment) PostEquipmentFunc(repository repositories.EquipmentRepository) equipment.CreateNewEquipmentHandlerFunc {
	return func(s equipment.CreateNewEquipmentParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		eq, err := repository.CreateEquipment(ctx, *s.NewEquipment)
		if err != nil {
			c.logger.Error("Error while creating equipment", zap.Error(err))
			return equipment.NewCreateNewEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while creating equipment"))
		}
		returnEq, err := mapEquipmentResponse(eq)
		if err != nil {
			c.logger.Error("Error while mapping equipment", zap.Error(err))
			return equipment.NewCreateNewEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while mapping equipment"))
		}

		return equipment.NewCreateNewEquipmentCreated().WithPayload(returnEq)
	}
}

func (c Equipment) GetEquipmentFunc(repository repositories.EquipmentRepository) equipment.GetEquipmentHandlerFunc {
	return func(s equipment.GetEquipmentParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		eq, err := repository.EquipmentByID(ctx, int(s.EquipmentID))
		if err != nil {
			c.logger.Error("Error while getting equipment", zap.Error(err))
			return equipment.NewGetEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while getting equipment"))
		}
		returnEq, err := mapEquipmentResponse(eq)
		if err != nil {
			c.logger.Error("Error while mapping equipment", zap.Error(err))
			return equipment.NewGetEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while mapping equipment"))
		}
		return equipment.NewGetEquipmentOK().WithPayload(returnEq)
	}
}

func (c Equipment) DeleteEquipmentFunc(repository repositories.EquipmentRepository,
	filesManager services.FileManager) equipment.DeleteEquipmentHandlerFunc {
	return func(s equipment.DeleteEquipmentParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		eq, err := repository.EquipmentByID(ctx, int(s.EquipmentID))
		if err != nil {
			c.logger.Error("Error while getting equipment", zap.Error(err))
			return equipment.NewDeleteEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: "Error while getting equipment",
				},
			})
		}
		err = repository.DeleteEquipmentByID(ctx, int(s.EquipmentID))
		if err != nil {
			c.logger.Error("Error while deleting equipment", zap.Error(err))
			return equipment.NewDeleteEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while deleting equipment"))
		}

		if err := repository.DeleteEquipmentPhoto(ctx, eq.Edges.Photo.ID); err != nil {
			c.logger.Error("Error while deleting photo from db", zap.Error(err))
		} else {
			err = filesManager.DeleteFile(eq.Edges.Photo.FileName)
			if err != nil {
				c.logger.Error("Error while deleting photo file", zap.Error(err))
			}
		}

		return equipment.NewDeleteEquipmentOK().WithPayload("Equipment deleted")
	}
}

func (c Equipment) ListEquipmentFunc(repository repositories.EquipmentRepository) equipment.GetAllEquipmentHandlerFunc {
	return func(s equipment.GetAllEquipmentParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		equipments, err := repository.AllEquipments(ctx)
		if err != nil {
			c.logger.Error("Error while getting all equipments", zap.Error(err))
			return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while getting all equipments"))
		}
		if len(equipments) == 0 {
			c.logger.Error("No equipments found")
			return equipment.NewGetAllEquipmentDefault(http.StatusNotFound).
				WithPayload(buildStringPayload("No equipments found"))
		}
		listEquipment := make([]*models.EquipmentResponse, len(equipments))
		for i, eq := range equipments {
			tmpEq, errMap := mapEquipmentResponse(eq)
			if errMap != nil {
				c.logger.Error("Error while mapping equipment", zap.Error(errMap))
				return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Error while mapping equipment"))
			}
			listEquipment[i] = tmpEq
		}
		return equipment.NewGetAllEquipmentOK().WithPayload(listEquipment)
	}
}

func (c Equipment) EditEquipmentFunc(repository repositories.EquipmentRepository) equipment.EditEquipmentHandlerFunc {
	return func(s equipment.EditEquipmentParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		eq, err := repository.UpdateEquipmentByID(ctx, int(s.EquipmentID), s.EditEquipment)
		if err != nil {
			c.logger.Error("Error while updating equipment", zap.Error(err))
			return equipment.NewEditEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while updating equipment"))
		}
		returnEq, err := mapEquipmentResponse(eq)
		if err != nil {
			c.logger.Error("Error while mapping equipment", zap.Error(err))
			return equipment.NewEditEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while mapping equipment"))
		}

		return equipment.NewEditEquipmentOK().WithPayload(returnEq)
	}
}

func (c Equipment) FindEquipmentFunc(EquipmentRepo repositories.EquipmentRepository) equipment.FindEquipmentHandlerFunc {
	return func(s equipment.FindEquipmentParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		equipmentFilter := *s.FindEquipment
		foundEquipment, err := EquipmentRepo.EquipmentsByFilter(ctx, equipmentFilter)
		if err != nil {
			c.logger.Error("Error while finding equipment", zap.Error(err))
			return equipment.NewFindEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while finding equipment"))
		}
		if len(foundEquipment) == 0 {
			c.logger.Info("Equipments not found")
			return equipment.NewFindEquipmentDefault(http.StatusNotFound).
				WithPayload(buildStringPayload("Equipments not found"))
		}
		returnEquipment := make([]*models.EquipmentResponse, len(foundEquipment))
		for i, eq := range foundEquipment {
			tmpEq, errMap := mapEquipmentResponse(eq)
			if errMap != nil {
				c.logger.Error("Error while mapping equipment", zap.Error(errMap))
				return equipment.NewFindEquipmentDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Error while mapping equipment"))
			}
			returnEquipment[i] = tmpEq
		}
		return equipment.NewFindEquipmentOK().WithPayload(returnEquipment)
	}
}

func mapEquipmentResponse(eq *ent.Equipment) (*models.EquipmentResponse, error) {
	if eq == nil {
		return nil, errors.New("equipment is nil")
	}
	id := int64(eq.ID)
	if eq.Edges.Kind == nil {
		return nil, errors.New("equipment kind is nil")
	}
	kindID := int64(eq.Edges.Kind.ID)
	if eq.Edges.Status == nil {
		return nil, errors.New("equipment status is nil")
	}
	statusID := int64(eq.Edges.Status.ID)

	var petKinds []*models.PetKind
	for _, petKindEdge := range eq.Edges.PetKinds {
		petKindID := int64(petKindEdge.ID)
		petKind := models.PetKind{ID: petKindID, Name: &petKindEdge.Name}
		petKinds = append(petKinds, &petKind)
	}

	var petSizeID *int64
	if eq.Edges.PetSize != nil {
		idInt64 := int64(eq.Edges.PetSize.ID)
		petSizeID = &idInt64
	}

	if eq.Edges.Photo == nil {
		return nil, errors.New("equipment photo is nil")
	}
	photoURL := eq.Edges.Photo.URL

	return &models.EquipmentResponse{
		Category:         &eq.Category,
		CompensationСost: &eq.CompensationCost,
		Condition:        &eq.Condition,
		Description:      &eq.Description,
		ID:               &id,
		InventoryNumber:  &eq.InventoryNumber,
		Kind:             &kindID,
		MaximumAmount:    &eq.MaximumAmount,
		MaximumDays:      &eq.MaximumDays,
		Name:             &eq.Name,
		ReceiptDate:      &eq.ReceiptDate,
		Status:           &statusID,
		Supplier:         &eq.Supplier,
		Title:            &eq.Title,
		PetSize:          petSizeID,
		Photo:            &photoURL,
		PetKinds:         petKinds,
	}, nil
}
