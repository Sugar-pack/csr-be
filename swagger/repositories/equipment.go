package repositories

import (
	"context"

	"entgo.io/ent/dialect/sql"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/predicate"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/statuses"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type EquipmentRepository interface {
	EquipmentsByFilter(ctx context.Context, filter models.EquipmentFilter) ([]*ent.Equipment, error)
}

type equipmentRepository struct {
	client *ent.Client
}

func OptionalIntEquipment(v int64, field string) predicate.Equipment {
	if v == 0 {
		return func(s *sql.Selector) {
		}
	}
	return func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(field), v))
	}
}

func OptionalIntStatus(v int64, field string) predicate.Statuses {
	if v == 0 {
		return func(s *sql.Selector) {
		}
	}
	return func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(field), v))
	}
}

func OptionalIntKind(v int64, field string) predicate.Kind {
	if v == 0 {
		return func(s *sql.Selector) {
		}
	}
	return func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(field), v))
	}
}

func OptionalStringEquipment(str string, field string) predicate.Equipment {
	if str == "" {
		return func(s *sql.Selector) {
		}
	}
	return func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(field), str))
	}
}

func (e *equipmentRepository) EquipmentsByFilter(ctx context.Context, filter models.EquipmentFilter) ([]*ent.Equipment, error) {
	result, err := e.client.Equipment.Query().
		QueryStatus().
		Where(OptionalIntStatus(filter.Status, statuses.FieldID)).
		QueryEquipments().
		QueryKind().
		Where(OptionalIntKind(filter.Kind, kind.FieldID)).
		QueryEquipments().
		Where(
			equipment.NameContains(filter.NameSubstring),
			OptionalStringEquipment(filter.Name, equipment.FieldName),
			OptionalStringEquipment(filter.Description, equipment.FieldDescription),
			OptionalStringEquipment(filter.Category, equipment.FieldCategory),
			OptionalIntEquipment(filter.CompensationСost, equipment.FieldCompensationCost),
			OptionalStringEquipment(filter.Condition, equipment.FieldCondition),
			OptionalIntEquipment(filter.InventoryNumber, equipment.FieldInventoryNumber),
			OptionalStringEquipment(filter.Supplier, equipment.FieldSupplier),
			OptionalStringEquipment(filter.ReceiptDate, equipment.FieldReceiptDate),
			OptionalIntEquipment(filter.MaximumAmount, equipment.FieldMaximumAmount),
			OptionalIntEquipment(filter.MaximumDays, equipment.FieldMaximumDays),
		).
		WithKind().
		WithStatus().
		All(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewEquipmentRepository(client *ent.Client) EquipmentRepository {
	return &equipmentRepository{
		client: client,
	}
}
