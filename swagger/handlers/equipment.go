package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/status"
)

type Equipment struct {
	client *ent.Client
	logger *zap.Logger
}

func NewEquipment(client *ent.Client, logger *zap.Logger) *Equipment {
	return &Equipment{
		client: client,
		logger: logger,
	}
}

func (c Equipment) PostEquipmentFunc() equipment.CreateNewEquipmentHandlerFunc {
	return func(s equipment.CreateNewEquipmentParams) middleware.Responder {

		e, err := c.client.Equipment.Create().
			SetName(*s.NewEquipment.Name).
			SetDescription(*s.NewEquipment.Description).
			SetCategory(*s.NewEquipment.Category).
			SetCompensationCost(*s.NewEquipment.CompensationСost).
			SetCondition(*s.NewEquipment.Condition).
			SetInventoryNumber(*s.NewEquipment.InventoryNumber).
			SetSupplier(*s.NewEquipment.Supplier).
			SetReceiptDate(*s.NewEquipment.ReceiptDate).
			SetMaximumAmount(*s.NewEquipment.MaximumAmount).
			SetMaximumDays(*s.NewEquipment.MaximumDays).
			SetKind(&ent.Kind{ID: int(*s.NewEquipment.Kind)}).
			SetStatus(&ent.Statuses{ID: int(*s.NewEquipment.Status)}).
			Save(s.HTTPRequest.Context())

		if err != nil {
			return equipment.NewCreateNewEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		kind, err := e.QueryKind().Only(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewGetEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		kindId := int64(kind.ID)

		status, err := e.QueryStatus().Only(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewGetEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		statusId := int64(status.ID)

		id := int64(e.ID)
		return equipment.NewCreateNewEquipmentCreated().WithPayload(&models.EquipmentResponse{
			ID:               &id,
			Description:      &e.Description,
			Name:             &e.Name,
			Category:         &e.Category,
			CompensationСost: &e.CompensationCost,
			Condition:        &e.Condition,
			InventoryNumber:  &e.InventoryNumber,
			Supplier:         &e.Supplier,
			ReceiptDate:      &e.ReceiptDate,
			MaximumAmount:    &e.MaximumAmount,
			MaximumDays:      &e.MaximumDays,
			Kind:             &kindId,
			Status:           &statusId,
		})
	}
}

func (c Equipment) GetEquipmentFunc() equipment.GetEquipmentHandlerFunc {
	return func(s equipment.GetEquipmentParams) middleware.Responder {
		equipmentId, err := strconv.Atoi(s.EquipmentID)
		if err != nil {
			return equipment.NewGetEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		e, err := c.client.Equipment.Get(s.HTTPRequest.Context(), equipmentId)
		if err != nil {
			return equipment.NewGetEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		kind, err := e.QueryKind().Only(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewGetEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		kindId := int64(kind.ID)

		status, err := e.QueryStatus().Only(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewGetEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		statusId := int64(status.ID)
		return equipment.NewGetEquipmentCreated().WithPayload(&models.Equipment{
			Description:      &e.Description,
			Name:             &e.Name,
			Category:         &e.Category,
			CompensationСost: &e.CompensationCost,
			Condition:        &e.Condition,
			InventoryNumber:  &e.InventoryNumber,
			Supplier:         &e.Supplier,
			ReceiptDate:      &e.ReceiptDate,
			MaximumAmount:    &e.MaximumAmount,
			MaximumDays:      &e.MaximumDays,
			Kind:             &kindId,
			Status:           &statusId,
		})
	}
}

func (c Equipment) DeleteEquipmentFunc() equipment.DeleteEquipmentHandlerFunc {
	return func(s equipment.DeleteEquipmentParams) middleware.Responder {
		equipmentId, err := strconv.Atoi(s.EquipmentID)
		if err != nil {
			return status.NewDeleteStatusDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		e, err := c.client.Equipment.Get(s.HTTPRequest.Context(), equipmentId)
		if err != nil {
			return equipment.NewDeleteEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		kind, err := e.QueryKind().Only(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewDeleteEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		kindId := int64(kind.ID)

		status, err := e.QueryStatus().Only(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		statusId := int64(status.ID)

		forReturn := &models.Equipment{
			Description:      &e.Description,
			Name:             &e.Name,
			Category:         &e.Category,
			CompensationСost: &e.CompensationCost,
			Condition:        &e.Condition,
			InventoryNumber:  &e.InventoryNumber,
			Supplier:         &e.Supplier,
			ReceiptDate:      &e.ReceiptDate,
			MaximumAmount:    &e.MaximumAmount,
			MaximumDays:      &e.MaximumDays,
			Kind:             &kindId,
			Status:           &statusId,
		}
		err = c.client.Equipment.DeleteOne(e).Exec(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewDeleteEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		return equipment.NewDeleteEquipmentCreated().WithPayload(forReturn)
	}
}

func (c Equipment) ListEquipmentFunc() equipment.GetAllEquipmentHandlerFunc {
	return func(s equipment.GetAllEquipmentParams) middleware.Responder {
		e, err := c.client.Equipment.Query().All(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		listEquipment := models.ListEquipment{}
		for _, element := range e {

			id := int64(element.ID)

			kind, err := element.QueryKind().Only(s.HTTPRequest.Context())
			if err != nil {
				return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
					Data: &models.ErrorData{
						Message: err.Error(),
					},
				})
			}
			kindId := int64(kind.ID)

			status, err := element.QueryStatus().Only(s.HTTPRequest.Context())
			if err != nil {
				return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
					Data: &models.ErrorData{
						Message: err.Error(),
					},
				})
			}

			statusId := int64(status.ID)

			listEquipment = append(listEquipment, &models.EquipmentResponse{
				ID:               &id,
				Description:      &element.Description,
				Name:             &element.Name,
				Category:         &element.Category,
				CompensationСost: &element.CompensationCost,
				Condition:        &element.Condition,
				InventoryNumber:  &element.InventoryNumber,
				Supplier:         &element.Supplier,
				ReceiptDate:      &element.ReceiptDate,
				MaximumAmount:    &element.MaximumAmount,
				MaximumDays:      &element.MaximumDays,
				Kind:             &kindId,
				Status:           &statusId,
			})
		}
		return equipment.NewGetAllEquipmentCreated().WithPayload(listEquipment)
	}
}

func (c Equipment) EditEquipmentFunc() equipment.EditEquipmentHandlerFunc {
	return func(s equipment.EditEquipmentParams) middleware.Responder {
		equipmentId, err := strconv.Atoi(s.EquipmentID)
		if err != nil {
			return equipment.NewEditEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		e, err := c.client.Equipment.Get(s.HTTPRequest.Context(), equipmentId)
		if err != nil {
			return equipment.NewEditEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		edit := e.Update()
		if *s.EditEquipment.Name != "" {
			edit.SetName(*s.EditEquipment.Name)
		}
		if *s.EditEquipment.Category != "" {
			edit.SetCategory(*s.EditEquipment.Category)
		}
		if *s.EditEquipment.Description != "" {
			edit.SetDescription(*s.EditEquipment.Description)
		}
		if *s.EditEquipment.CompensationСost != 0 {
			edit.SetCompensationCost(*s.EditEquipment.CompensationСost)
		}
		if *s.EditEquipment.Condition != "" {
			edit.SetCondition(*s.EditEquipment.Condition)
		}
		if *s.EditEquipment.InventoryNumber != 0 {
			edit.SetInventoryNumber(*s.EditEquipment.InventoryNumber)
		}
		if *s.EditEquipment.Supplier != "" {
			edit.SetSupplier(*s.EditEquipment.Supplier)
		}
		if *s.EditEquipment.ReceiptDate != "" {
			edit.SetReceiptDate(*s.EditEquipment.ReceiptDate)
		}
		if *s.EditEquipment.MaximumAmount != 0 {
			edit.SetMaximumAmount(*s.EditEquipment.MaximumAmount)
		}
		if *s.EditEquipment.MaximumDays != 0 {
			edit.SetMaximumDays(*s.EditEquipment.MaximumDays)
		}
		if *s.EditEquipment.Kind != 0 {
			edit.SetKind(&ent.Kind{ID: int(*s.EditEquipment.Kind)})
		}

		if *s.EditEquipment.Status != 0 {
			edit.SetStatus(&ent.Statuses{ID: int(*s.EditEquipment.Status)})
		}
		res, err := edit.Save(s.HTTPRequest.Context())
		//res, err := c.client.Equipment.Get(s.HTTPRequest.Context(), equipmentId)
		if err != nil {
			return equipment.NewEditEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}

		kind, err := res.QueryKind().Only(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		kindId := int64(kind.ID)

		status, err := res.QueryStatus().Only(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		statusId := int64(status.ID)

		return equipment.NewEditEquipmentCreated().WithPayload(&models.Equipment{
			Description:      &res.Description,
			Name:             &res.Name,
			Category:         &res.Category,
			CompensationСost: &res.CompensationCost,
			Condition:        &res.Condition,
			InventoryNumber:  &res.InventoryNumber,
			Supplier:         &res.Supplier,
			ReceiptDate:      &res.ReceiptDate,
			MaximumAmount:    &res.MaximumAmount,
			MaximumDays:      &res.MaximumDays,
			Kind:             &kindId,
			Status:           &statusId,
		})
	}
}

func (c Equipment) FindEquipmentFunc() equipment.FindEquipmentHandlerFunc {
	return func(s equipment.FindEquipmentParams) middleware.Responder {
		all, err := c.client.Equipment.Query().All(s.HTTPRequest.Context())
		if err != nil {
			return equipment.NewFindEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
				Data: &models.ErrorData{
					Message: err.Error(),
				},
			})
		}
		res := models.ListEquipment{}
		for _, element := range all {

			kind, err := element.QueryKind().Only(s.HTTPRequest.Context())
			if err != nil {
				return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
					Data: &models.ErrorData{
						Message: err.Error(),
					},
				})
			}
			kindId := int64(kind.ID)

			status, err := element.QueryStatus().Only(s.HTTPRequest.Context())
			if err != nil {
				return equipment.NewGetAllEquipmentDefault(http.StatusInternalServerError).WithPayload(&models.Error{
					Data: &models.ErrorData{
						Message: err.Error(),
					},
				})
			}
			statusId := int64(status.ID)

			if (s.FindEquipment.Name == "" || element.Name == s.FindEquipment.Name) &&
				(s.FindEquipment.NameSubstring == "" || strings.Contains(strings.ToLower(element.Name), strings.ToLower(s.FindEquipment.NameSubstring))) &&
				(s.FindEquipment.Description == "" || element.Description == s.FindEquipment.Description) &&
				(s.FindEquipment.Category == "" || element.Category == s.FindEquipment.Category) &&
				(s.FindEquipment.CompensationСost == 0 || element.CompensationCost == s.FindEquipment.CompensationСost) &&
				(s.FindEquipment.Condition == "" || element.Condition == s.FindEquipment.Condition) &&
				(s.FindEquipment.InventoryNumber == 0 || element.InventoryNumber == s.FindEquipment.InventoryNumber) &&
				(s.FindEquipment.Supplier == "" || element.Supplier == s.FindEquipment.Supplier) &&
				(s.FindEquipment.ReceiptDate == "" || element.ReceiptDate == s.FindEquipment.ReceiptDate) &&
				(s.FindEquipment.MaximumAmount == 0 || element.MaximumAmount == s.FindEquipment.MaximumAmount) &&
				(s.FindEquipment.MaximumDays == 0 || element.MaximumDays == s.FindEquipment.MaximumDays) &&
				(s.FindEquipment.Kind == 0 || kindId == s.FindEquipment.Kind) &&
				(s.FindEquipment.Status == 0 || statusId == s.FindEquipment.Status) {

				id := int64(element.ID)
				res = append(res, &models.EquipmentResponse{
					ID:               &id,
					Description:      &element.Description,
					Name:             &element.Name,
					Category:         &element.Category,
					CompensationСost: &element.CompensationCost,
					Condition:        &element.Condition,
					InventoryNumber:  &element.InventoryNumber,
					Supplier:         &element.Supplier,
					ReceiptDate:      &element.ReceiptDate,
					MaximumAmount:    &element.MaximumAmount,
					MaximumDays:      &element.MaximumDays,
					Kind:             &kindId,
					Status:           &statusId,
				})
			}
		}
		return equipment.NewFindEquipmentCreated().WithPayload(res)
	}
}
