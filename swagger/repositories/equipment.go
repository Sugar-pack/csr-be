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
	CreateEquipment(ctx context.Context, eq models.Equipment) (*ent.Equipment, error)
	EquipmentByID(ctx context.Context, id int) (*ent.Equipment, error)
	DeleteEquipmentByID(ctx context.Context, id int) error
	AllEquipments(ctx context.Context) ([]*ent.Equipment, error)
	UpdateEquipmentByID(ctx context.Context, id int, eq *models.Equipment) (*ent.Equipment, error)
}

type equipmentRepository struct {
	client *ent.Client
}

func NewEquipmentRepository(client *ent.Client) EquipmentRepository {
	return &equipmentRepository{
		client: client,
	}
}

func (r *equipmentRepository) EquipmentsByFilter(ctx context.Context, filter models.EquipmentFilter) ([]*ent.Equipment, error) {
	result, err := r.client.Equipment.Query().
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
			OptionalStringEquipment(filter.Title, equipment.FieldTitle),
		).
		WithKind().
		WithStatus().
		WithPetKinds().
		All(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *equipmentRepository) CreateEquipment(ctx context.Context, NewEquipment models.Equipment) (*ent.Equipment, error) {
	var petKinds []int
	for _, id := range NewEquipment.PetKinds {
		petKinds = append(petKinds, int(id))
	}

	eq, err := r.client.Equipment.Create().
		SetName(*NewEquipment.Name).
		SetDescription(*NewEquipment.Description).
		SetCategory(*NewEquipment.Category).
		SetCompensationCost(*NewEquipment.CompensationСost).
		SetCondition(*NewEquipment.Condition).
		SetInventoryNumber(*NewEquipment.InventoryNumber).
		SetSupplier(*NewEquipment.Supplier).
		SetReceiptDate(*NewEquipment.ReceiptDate).
		SetMaximumAmount(*NewEquipment.MaximumAmount).
		SetMaximumDays(*NewEquipment.MaximumDays).
		SetKind(&ent.Kind{ID: int(*NewEquipment.Kind)}).
		SetStatus(&ent.Statuses{ID: int(*NewEquipment.Status)}).
		SetKindID(int(*NewEquipment.Kind)).
		SetStatusID(int(*NewEquipment.Status)).
		AddPetKindIDs(petKinds...).
		SetTitle(*NewEquipment.Title).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	result, err := r.client.Equipment.Query().Where(equipment.ID(eq.ID)).WithKind().WithStatus().WithPetKinds().Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *equipmentRepository) EquipmentByID(ctx context.Context, id int) (*ent.Equipment, error) {
	result, err := r.client.Equipment.Query().Where(equipment.ID(id)).WithKind().WithStatus().WithPetKinds().Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *equipmentRepository) DeleteEquipmentByID(ctx context.Context, id int) error {
	_, err := r.client.Equipment.Delete().Where(equipment.ID(id)).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *equipmentRepository) AllEquipments(ctx context.Context) ([]*ent.Equipment, error) {
	result, err := r.client.Equipment.Query().WithKind().WithStatus().WithPetKinds().All(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *equipmentRepository) UpdateEquipmentByID(ctx context.Context, id int, eq *models.Equipment) (*ent.Equipment, error) {
	eqToUpdate, err := r.client.Equipment.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	edit := eqToUpdate.Update()
	if *eq.Name != "" {
		edit.SetName(*eq.Name)
	}
	if *eq.Category != "" {
		edit.SetCategory(*eq.Category)
	}
	if *eq.Description != "" {
		edit.SetDescription(*eq.Description)
	}
	if *eq.CompensationСost != 0 {
		edit.SetCompensationCost(*eq.CompensationСost)
	}
	if *eq.Condition != "" {
		edit.SetCondition(*eq.Condition)
	}
	if *eq.InventoryNumber != 0 {
		edit.SetInventoryNumber(*eq.InventoryNumber)
	}
	if *eq.Supplier != "" {
		edit.SetSupplier(*eq.Supplier)
	}
	if *eq.ReceiptDate != "" {
		edit.SetReceiptDate(*eq.ReceiptDate)
	}
	if *eq.MaximumAmount != 0 {
		edit.SetMaximumAmount(*eq.MaximumAmount)
	}
	if *eq.MaximumDays != 0 {
		edit.SetMaximumDays(*eq.MaximumDays)
	}
	if *eq.Kind != 0 {
		edit.SetKind(&ent.Kind{ID: int(*eq.Kind)})
	}
	if pks := []int{}; len(eq.PetKinds) != 0 {
		for _, petKind := range eq.PetKinds {
			pks = append(pks, int(petKind))
		}
		edit.AddPetKindIDs(pks...)
	}
	if *eq.Title != "" {
		edit.SetTitle(*eq.Title)
	}

	if *eq.Status != 0 {
		edit.SetStatus(&ent.Statuses{ID: int(*eq.Status)})
	}
	_, err = edit.Save(ctx)
	if err != nil {
		return nil, err
	}
	result, err := r.client.Equipment.Query().Where(equipment.ID(eqToUpdate.ID)).WithKind().WithStatus().Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
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
