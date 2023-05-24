package handlers

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/restapi/operations/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func SetEquipmentHandler(logger *zap.Logger, api *operations.BeAPI) {
	eqRepo := repositories.NewEquipmentRepository()
	eqStatusNameRepo := repositories.NewEquipmentStatusNameRepository()
	equipmentHandler := NewEquipment(logger)
	api.EquipmentCreateNewEquipmentHandler = equipmentHandler.PostEquipmentFunc(eqRepo, eqStatusNameRepo)
	api.EquipmentGetEquipmentHandler = equipmentHandler.GetEquipmentFunc(eqRepo)
	api.EquipmentDeleteEquipmentHandler = equipmentHandler.DeleteEquipmentFunc(eqRepo)
	api.EquipmentGetAllEquipmentHandler = equipmentHandler.ListEquipmentFunc(eqRepo)
	api.EquipmentEditEquipmentHandler = equipmentHandler.EditEquipmentFunc(eqRepo)
	api.EquipmentFindEquipmentHandler = equipmentHandler.FindEquipmentFunc(eqRepo)
	api.EquipmentArchiveEquipmentHandler = equipmentHandler.ArchiveEquipmentFunc(eqRepo)
}

type Equipment struct {
	logger *zap.Logger
}

const EquipmentNotFoundMsg = "Equipment not found"

func NewEquipment(logger *zap.Logger) *Equipment {
	return &Equipment{
		logger: logger,
	}
}

func (c Equipment) PostEquipmentFunc(eqRepo domain.EquipmentRepository, eqStatusNameRepo domain.EquipmentStatusNameRepository) equipment.CreateNewEquipmentHandlerFunc {
	return func(s equipment.CreateNewEquipmentParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		status, err := eqStatusNameRepo.GetByName(ctx, "available")
		if err != nil {
			c.logger.Error("Error while getting status", zap.Error(err))
			return equipment.NewCreateNewEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while creating equipment"))
		}
		eq, err := eqRepo.CreateEquipment(ctx, *s.NewEquipment, status)
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

func (c Equipment) GetEquipmentFunc(repository domain.EquipmentRepository) equipment.GetEquipmentHandlerFunc {
	return func(s equipment.GetEquipmentParams, access interface{}) middleware.Responder {
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

func (c Equipment) ArchiveEquipmentFunc(repository domain.EquipmentRepository) equipment.ArchiveEquipmentHandlerFunc {
	return func(s equipment.ArchiveEquipmentParams, _ interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		c.logger.Info("Archive equipment", zap.Int64("equipmentID", s.EquipmentID))
		err := repository.ArchiveEquipment(ctx, int(s.EquipmentID))
		if err != nil {
			if ent.IsNotFound(err) {
				return equipment.NewArchiveEquipmentNotFound().
					WithPayload(buildStringPayload(EquipmentNotFoundMsg))
			}
			c.logger.Error("Error while archiving equipment", zap.Error(err))
			return equipment.NewArchiveEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while archiving equipment"))
		}
		return equipment.NewArchiveEquipmentNoContent()
	}
}

func (c Equipment) DeleteEquipmentFunc(repository domain.EquipmentRepository) equipment.DeleteEquipmentHandlerFunc {
	return func(s equipment.DeleteEquipmentParams, access interface{}) middleware.Responder {
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
		}

		return equipment.NewDeleteEquipmentOK().WithPayload("Equipment deleted")
	}
}

func (c Equipment) ListEquipmentFunc(repository domain.EquipmentRepository) equipment.GetAllEquipmentHandlerFunc {
	return func(s equipment.GetAllEquipmentParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		limit := utils.GetValueByPointerOrDefaultValue(s.Limit, math.MaxInt)
		offset := utils.GetValueByPointerOrDefaultValue(s.Offset, 0)
		orderBy := utils.GetValueByPointerOrDefaultValue(s.OrderBy, utils.AscOrder)
		orderColumn := utils.GetValueByPointerOrDefaultValue(s.OrderColumn, order.FieldID)
		total, err := repository.AllEquipmentsTotal(ctx)
		if err != nil {
			c.logger.Error("Error while getting total of all equipments", zap.Error(err))
			return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while getting total of all equipments"))
		}
		var equipments []*ent.Equipment
		if total > 0 {
			equipments, err = repository.AllEquipments(ctx, int(limit), int(offset), orderBy, orderColumn)
			if err != nil {
				c.logger.Error("Error while getting all equipments", zap.Error(err))
				return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Error while getting all equipments"))
			}
		}
		totalEquipments := int64(total)
		listEquipment := &models.ListEquipment{
			Items: make([]*models.EquipmentResponse, len(equipments)),
			Total: &totalEquipments,
		}
		for i, eq := range equipments {
			tmpEq, errMap := mapEquipmentResponse(eq)
			if errMap != nil {
				c.logger.Error("Error while mapping equipment", zap.Error(errMap))
				return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Error while mapping equipment"))
			}
			listEquipment.Items[i] = tmpEq
		}
		return equipment.NewGetAllEquipmentOK().WithPayload(listEquipment)
	}
}

func (c Equipment) EditEquipmentFunc(repository domain.EquipmentRepository) equipment.EditEquipmentHandlerFunc {
	return func(s equipment.EditEquipmentParams, access interface{}) middleware.Responder {
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

func (c Equipment) FindEquipmentFunc(repository domain.EquipmentRepository) equipment.FindEquipmentHandlerFunc {
	return func(s equipment.FindEquipmentParams, access interface{}) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		limit := utils.GetValueByPointerOrDefaultValue(s.Limit, math.MaxInt)
		offset := utils.GetValueByPointerOrDefaultValue(s.Offset, 0)
		orderBy := utils.GetValueByPointerOrDefaultValue(s.OrderBy, utils.AscOrder)
		orderColumn := utils.GetValueByPointerOrDefaultValue(s.OrderColumn, order.FieldID)
		equipmentFilter := *s.FindEquipment
		total, err := repository.EquipmentsByFilterTotal(ctx, equipmentFilter)
		if err != nil {
			c.logger.Error("Error while getting total of all equipments", zap.Error(err))
			return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).
				WithPayload(buildStringPayload("Error while getting total of all equipments"))
		}
		var foundEquipment []*ent.Equipment
		if total > 0 {
			foundEquipment, err = repository.EquipmentsByFilter(ctx, equipmentFilter, int(limit), int(offset), orderBy, orderColumn)
			if err != nil {
				c.logger.Error("Error while finding equipment", zap.Error(err))
				return equipment.NewFindEquipmentDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Error while finding equipment"))
			}
		}
		totalEquipments := int64(total)
		returnEquipment := &models.ListEquipment{
			Items: make([]*models.EquipmentResponse, len(foundEquipment)),
			Total: &totalEquipments,
		}
		for i, eq := range foundEquipment {
			tmpEq, errMap := mapEquipmentResponse(eq)
			if errMap != nil {
				c.logger.Error("Error while mapping equipment", zap.Error(errMap))
				return equipment.NewFindEquipmentDefault(http.StatusInternalServerError).
					WithPayload(buildStringPayload("Error while mapping equipment"))
			}
			returnEquipment.Items[i] = tmpEq
		}
		return equipment.NewFindEquipmentOK().WithPayload(returnEquipment)
	}
}

func mapEquipmentResponse(eq *ent.Equipment) (*models.EquipmentResponse, error) {
	if eq == nil {
		return nil, errors.New("equipment is nil")
	}
	id := int64(eq.ID)
	if eq.Edges.Category == nil {
		return nil, errors.New("equipment category is nil")
	}
	categoryID := int64(eq.Edges.Category.ID)
	subcategoryID := int64(0)
	if eq.Edges.Subcategory != nil && eq.Edges.Subcategory.ID > 0 {
		subcategoryID = int64(eq.Edges.Subcategory.ID)
	}
	if eq.Edges.CurrentStatus == nil {
		return nil, errors.New("equipment status is nil")
	}
	statusID := int64(eq.Edges.CurrentStatus.ID)

	petKinds := make([]*models.PetKind, len(eq.Edges.PetKinds))
	for i, petKindEdge := range eq.Edges.PetKinds {
		petKind := models.PetKind{Name: &petKindEdge.Name}
		petKinds[i] = &petKind
	}

	var petSizeID *int64
	if eq.Edges.PetSize != nil {
		idInt64 := int64(eq.Edges.PetSize.ID)
		petSizeID = &idInt64
	}

	var photoID string
	if eq.Edges.Photo != nil {
		photoID = eq.Edges.Photo.ID
	}

	var eqReceiptDate int64
	if eq.ReceiptDate != "" {
		eqReceiptTime, err := time.Parse(utils.TimeFormat, eq.ReceiptDate)
		if err != nil {
			return nil, err
		}

		eqReceiptDate = eqReceiptTime.Unix()
	}

	return &models.EquipmentResponse{
		TermsOfUse:       &eq.TermsOfUse,
		CompensationCost: &eq.CompensationCost,
		TechnicalIssues:  &eq.TechIssue,
		Condition:        eq.Condition,
		Description:      &eq.Description,
		ID:               &id,
		InventoryNumber:  &eq.InventoryNumber,
		Category:         &categoryID,
		Subcategory:      subcategoryID,
		MaximumDays:      &eq.MaximumDays,
		Name:             &eq.Name,
		ReceiptDate:      &eqReceiptDate,
		Status:           &statusID,
		Supplier:         &eq.Supplier,
		Title:            &eq.Title,
		PetSize:          petSizeID,
		PhotoID:          &photoID,
		PetKinds:         petKinds,
	}, nil
}
