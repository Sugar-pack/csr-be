package repositories

import (
	"context"
	"errors"

	"entgo.io/ent/dialect/sql"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipmentstatusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/petkind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/petsize"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/photo"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/predicate"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/subcategory"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

var fieldsToOrderEquipments = []string{
	equipment.FieldID,
	equipment.FieldName,
	equipment.FieldTitle,
}

type equipmentRepository struct {
}

func NewEquipmentRepository() domain.EquipmentRepository {
	return &equipmentRepository{}
}

func (r *equipmentRepository) EquipmentsByFilter(ctx context.Context, filter models.EquipmentFilter,
	limit, offset int, orderBy, orderColumn string) ([]*ent.Equipment, error) {
	if !utils.IsValueInList(orderColumn, fieldsToOrderEquipments) {
		return nil, errors.New("wrong column to order by")
	}
	orderFunc, err := utils.GetOrderFunc(orderBy, orderColumn)
	if err != nil {
		return nil, err
	}
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}

	result, err := tx.Equipment.Query().
		QueryCurrentStatus().
		Where(OptionalIntStatus(filter.Status, equipmentstatusname.FieldID)).
		QueryEquipments().
		QueryCategory().
		Where(OptionalIntCategory(filter.Category, category.FieldID)).
		QueryEquipments().
		QuerySubcategory().
		Where(OptionalIntSubcategory(filter.Subcategory, subcategory.FieldID)).
		QueryEquipments().
		Where(
			equipment.NameContains(filter.NameSubstring),
			OptionalStringEquipment(filter.Name, equipment.FieldName),
			OptionalStringEquipment(filter.Description, equipment.FieldDescription),
			OptionalStringEquipment(filter.TermsOfUse, equipment.FieldTermsOfUse),
			OptionalIntEquipment(filter.CompensationCost, equipment.FieldCompensationCost),
			OptionalIntEquipment(filter.InventoryNumber, equipment.FieldInventoryNumber),
			OptionalStringEquipment(filter.Supplier, equipment.FieldSupplier),
			OptionalStringEquipment(filter.ReceiptDate, equipment.FieldReceiptDate),
			OptionalStringEquipment(filter.Title, equipment.FieldTitle),
			OptionalBoolEquipment(filter.TechnicalIssues, equipment.FieldTechIssue),
			OptionalStringEquipment(filter.Condition, equipment.FieldCondition),
			OptionalIntEquipment(filter.MaximumAmount, equipment.FieldMaximumAmount),
			OptionalIntEquipment(filter.MaximumDays, equipment.FieldMaximumDays),
			equipment.HasPetKindsWith(OptionalIntsPetKind(filter.PetKinds, petkind.FieldID)),
			equipment.HasPetSizeWith(OptionalIntsPetSize(filter.PetSize, petsize.FieldID)),
		).
		Order(orderFunc).
		Limit(limit).Offset(offset).
		WithPetSize().
		WithCategory().
		WithSubcategory().
		WithCurrentStatus().
		WithPhoto().
		WithPetKinds().
		All(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *equipmentRepository) CreateEquipment(ctx context.Context, NewEquipment models.Equipment, status *ent.EquipmentStatusName) (*ent.Equipment, error) {
	var petKinds []int
	for _, id := range NewEquipment.PetKinds {
		petKinds = append(petKinds, int(id))
	}
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	eq, err := tx.Equipment.Create().
		SetName(*NewEquipment.Name).
		SetDescription(*NewEquipment.Description).
		SetTermsOfUse(NewEquipment.TermsOfUse).
		SetCompensationCost(*NewEquipment.CompensationCost).
		SetTechIssue(*NewEquipment.TechnicalIssues).
		SetCondition(NewEquipment.Condition).
		SetInventoryNumber(*NewEquipment.InventoryNumber).
		SetSupplier(*NewEquipment.Supplier).
		SetReceiptDate(*NewEquipment.ReceiptDate).
		SetMaximumAmount(*NewEquipment.MaximumAmount).
		SetMaximumDays(*NewEquipment.MaximumDays).
		SetCategory(&ent.Category{ID: int(*NewEquipment.Category)}).
		SetCurrentStatus(status).
		SetCategoryID(int(*NewEquipment.Category)).
		SetSubcategoryID(int(NewEquipment.Subcategory)).
		SetSubcategory(&ent.Subcategory{ID: int(NewEquipment.Subcategory)}).
		SetCurrentStatusID(int(*NewEquipment.Status)).
		AddPetKindIDs(petKinds...).
		SetTitle(*NewEquipment.Title).
		SetPhoto(&ent.Photo{ID: *NewEquipment.PhotoID}).
		SetPetSizeID(int(*NewEquipment.PetSize)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	result, err := tx.Equipment.Query().Where(equipment.ID(eq.ID)).
		WithCategory().WithSubcategory().WithCurrentStatus().WithPhoto().WithPetKinds().WithPetSize().Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *equipmentRepository) EquipmentByID(ctx context.Context, id int) (*ent.Equipment, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := tx.Equipment.Query().Where(equipment.ID(id)).
		WithCategory().WithSubcategory().WithCurrentStatus().WithPetKinds().WithPetSize().WithPhoto().Only(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *equipmentRepository) DeleteEquipmentByID(ctx context.Context, id int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Equipment.Delete().Where(equipment.ID(id)).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *equipmentRepository) DeleteEquipmentPhoto(ctx context.Context, id string) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Photo.Delete().Where(photo.ID(id)).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *equipmentRepository) AllEquipments(ctx context.Context, limit, offset int, orderBy, orderColumn string) ([]*ent.Equipment, error) {
	if !utils.IsValueInList(orderColumn, fieldsToOrderEquipments) {
		return nil, errors.New("wrong column to order by")
	}
	orderFunc, err := utils.GetOrderFunc(orderBy, orderColumn)
	if err != nil {
		return nil, err
	}
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	result, err := tx.Equipment.Query().Order(orderFunc).Limit(limit).Offset(offset).
		WithCategory().WithSubcategory().WithCurrentStatus().WithPetKinds().WithPetSize().WithPhoto().
		All(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *equipmentRepository) AllEquipmentsTotal(ctx context.Context) (int, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return 0, err
	}
	total, err := tx.Equipment.Query().
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *equipmentRepository) EquipmentsByFilterTotal(ctx context.Context, filter models.EquipmentFilter) (int, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return 0, err
	}

	total, err := tx.Equipment.Query().
		QueryCurrentStatus().
		Where(OptionalIntStatus(filter.Status, equipmentstatusname.FieldID)).
		QueryEquipments().
		QueryCategory().
		Where(OptionalIntCategory(filter.Category, category.FieldID)).
		QueryEquipments().
		QuerySubcategory().
		Where(OptionalIntSubcategory(filter.Subcategory, subcategory.FieldID)).
		QueryEquipments().
		Where(
			equipment.NameContains(filter.NameSubstring),
			OptionalStringEquipment(filter.Name, equipment.FieldName),
			OptionalStringEquipment(filter.Description, equipment.FieldDescription),
			OptionalStringEquipment(filter.TermsOfUse, equipment.FieldTermsOfUse),
			OptionalIntEquipment(filter.CompensationCost, equipment.FieldCompensationCost),
			OptionalIntEquipment(filter.InventoryNumber, equipment.FieldInventoryNumber),
			OptionalStringEquipment(filter.Supplier, equipment.FieldSupplier),
			OptionalStringEquipment(filter.ReceiptDate, equipment.FieldReceiptDate),
			OptionalStringEquipment(filter.Title, equipment.FieldTitle),
			OptionalBoolEquipment(filter.TechnicalIssues, equipment.FieldTechIssue),
			OptionalStringEquipment(filter.Condition, equipment.FieldCondition),
			OptionalIntEquipment(filter.MaximumAmount, equipment.FieldMaximumAmount),
			OptionalIntEquipment(filter.MaximumDays, equipment.FieldMaximumDays),
			equipment.HasPetKindsWith(OptionalIntsPetKind(filter.PetKinds, petkind.FieldID)),
			equipment.HasPetSizeWith(OptionalIntsPetSize(filter.PetSize, petsize.FieldID)),
		).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *equipmentRepository) UpdateEquipmentByID(ctx context.Context, id int, eq *models.Equipment) (*ent.Equipment, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	eqToUpdate, err := tx.Equipment.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	edit := eqToUpdate.Update()
	if *eq.Name != "" {
		edit.SetName(*eq.Name)
	}
	if eq.TermsOfUse != "" {
		edit.SetTermsOfUse(eq.TermsOfUse)
	}
	if *eq.Description != "" {
		edit.SetDescription(*eq.Description)
	}
	if *eq.CompensationCost != 0 {
		edit.SetCompensationCost(*eq.CompensationCost)
	}
	if eq.TechnicalIssues != nil {
		edit.SetTechIssue(*eq.TechnicalIssues)
		edit.SetCondition(eq.Condition)
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
	if *eq.Category != 0 {
		edit.SetCategory(&ent.Category{ID: int(*eq.Category)})
	}
	if *eq.MaximumAmount != 0 {
		edit.SetMaximumAmount(*eq.MaximumAmount)
	}
	if *eq.MaximumDays != 0 {
		edit.SetMaximumDays(*eq.MaximumDays)
	}
	if eq.Subcategory != 0 {
		edit.SetSubcategory(&ent.Subcategory{ID: int(eq.Subcategory)})
	}
	if *eq.PetSize != 0 {
		edit.SetPetSizeID(int(*eq.PetSize))
	}
	edit.ClearPetKinds()
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
		edit.SetCurrentStatus(&ent.EquipmentStatusName{ID: int(*eq.Status)})
	}
	if *eq.PhotoID != "" {
		edit.SetPhoto(&ent.Photo{ID: *eq.PhotoID})
	}
	_, err = edit.Save(ctx)
	if err != nil {
		return nil, err
	}
	result, err := tx.Equipment.Query().Where(equipment.ID(eqToUpdate.ID)).
		WithCategory().WithSubcategory().WithCurrentStatus().WithPetSize().WithPetKinds().WithPhoto().Only(ctx)
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

func OptionalIntStatus(v int64, field string) predicate.EquipmentStatusName {
	if v == 0 {
		return func(s *sql.Selector) {
		}
	}
	return func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(field), v))
	}
}

func OptionalIntCategory(v int64, field string) predicate.Category {
	if v == 0 {
		return func(s *sql.Selector) {
		}
	}
	return func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(field), v))
	}
}

func OptionalIntSubcategory(v int64, field string) predicate.Subcategory {
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
		s.Where(sql.EqualFold(s.C(field), str))
	}
}

func OptionalBoolEquipment(b *bool, field string) predicate.Equipment {
	if b == nil {
		return func(s *sql.Selector) {
		}
	}
	return func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(field), b))
	}
}

func OptionalIntsPetSize(p []int64, field string) predicate.PetSize {
	if len(p) == 0 {
		return func(s *sql.Selector) {
		}
	}

	petSize := make([]int, len(p))
	for i, v := range p {
		petSize[i] = int(v)
	}

	return func(s *sql.Selector) {
		s.Where(sql.InInts(s.C(field), petSize...))
	}
}

func OptionalIntsPetKind(k []int64, field string) predicate.PetKind {
	if len(k) == 0 {
		return func(s *sql.Selector) {
		}
	}

	petKind := make([]int, len(k))
	for i, v := range k {
		petKind[i] = int(v)
	}

	return func(s *sql.Selector) {
		s.Where(sql.InInts(s.C(field), petKind...))
	}
}
