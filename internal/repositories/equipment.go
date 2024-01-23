package repositories

import (
	"context"
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipmentstatus"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipmentstatusname"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/orderstatusname"
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

	var filterReceiptDate string
	if filter.ReceiptDate != 0 {
		filterReceiptDate = time.Unix(filter.ReceiptDate, 0).Format(utils.TimeFormat)
	}

	result, err := tx.Equipment.Query().
		Where(
			equipment.HasCategoryWith(OptionalIntCategory(filter.Category, category.FieldID)),
			equipment.HasSubcategoryWith(OptionalIntSubcategory(filter.Subcategory, subcategory.FieldID)),
			equipment.HasCurrentStatusWith(OptionalIntStatus(filter.Status, equipmentstatusname.FieldID)),
			equipment.NameContains(filter.NameSubstring),
			OptionalStringEquipment(filter.Name, equipment.FieldName),
			OptionalStringEquipment(filter.Description, equipment.FieldDescription),
			OptionalStringEquipment(filter.TermsOfUse, equipment.FieldTermsOfUse),
			OptionalIntEquipment(filter.CompensationCost, equipment.FieldCompensationCost),
			OptionalIntEquipment(filter.InventoryNumber, equipment.FieldInventoryNumber),
			OptionalStringEquipment(filter.Supplier, equipment.FieldSupplier),
			OptionalStringEquipment(
				filterReceiptDate,
				equipment.FieldReceiptDate,
			),
			OptionalStringEquipment(filter.Title, equipment.FieldTitle),
			OptionalBoolEquipment(filter.TechnicalIssues, equipment.FieldTechIssue),
			OptionalStringEquipment(filter.Condition, equipment.FieldCondition),
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

	eqReceiptDate := time.Unix(*NewEquipment.ReceiptDate, 0).Format(utils.TimeFormat)

	eq, err := tx.Equipment.Create().
		SetName(*NewEquipment.Name).
		SetDescription(*NewEquipment.Description).
		SetTermsOfUse(NewEquipment.TermsOfUse).
		SetCompensationCost(*NewEquipment.CompensationCost).
		SetTechIssue(*NewEquipment.TechnicalIssues).
		SetCondition(NewEquipment.Condition).
		SetInventoryNumber(*NewEquipment.InventoryNumber).
		SetSupplier(*NewEquipment.Supplier).
		SetReceiptDate(eqReceiptDate).
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
		WithCategory().WithSubcategory().WithCurrentStatus().WithPetKinds().WithPetSize().WithPhoto().
		WithEquipmentStatus(func(esq *ent.EquipmentStatusQuery) {
			esq.
				Where(equipmentstatus.EndDateGTE(time.Now())).
				Where(equipmentstatus.HasEquipmentStatusNameWith(
					equipmentstatusname.NameEQ(domain.EquipmentStatusNotAvailable),
				))
		}).Only(ctx)
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
		WithCategory().WithSubcategory().WithCurrentStatus().WithPetKinds().
		WithPetSize().WithPhoto().
		WithEquipmentStatus(func(esq *ent.EquipmentStatusQuery) {
			esq.
				Where(equipmentstatus.EndDateGTE(time.Now())).
				Where(equipmentstatus.HasEquipmentStatusNameWith(
					equipmentstatusname.NameEQ(domain.EquipmentStatusNotAvailable),
				))
		}).
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

func (r *equipmentRepository) ArchiveEquipment(ctx context.Context, id int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	// check if equipment exists
	_, err = tx.Equipment.Get(ctx, id)
	if err != nil {
		return err
	}
	// get equipment status archived id
	equipmentStatusArchived, err := tx.EquipmentStatusName.Query().Where(equipmentstatusname.Name("archived")).Only(ctx)
	if err != nil {
		return err
	}
	// get closed order status id
	orderStatusClosed, err := tx.OrderStatusName.
		Query().
		Where(orderstatusname.Status(domain.OrderStatusClosed)).
		Only(ctx)
	if err != nil {
		return err
	}
	// get equipment status and change it to archived
	equipmentStatuses, err := tx.EquipmentStatus.Query().QueryEquipments().Where(equipment.ID(id)).QueryEquipmentStatus().All(ctx)
	if err != nil {
		return err
	}
	// if this equipment is not in order, then change status to archived
	if len(equipmentStatuses) == 0 {
		_, err = tx.Equipment.UpdateOneID(id).SetCurrentStatus(equipmentStatusArchived).Save(ctx)
		return err
	}
	// if this equipment is in order, then archive equipment status
	for _, equipmentStatus := range equipmentStatuses {
		_, err = equipmentStatus.Update().SetEquipmentStatusNameID(equipmentStatusArchived.ID).Save(ctx)
		if err != nil {
			return err
		}
		// get orders with this equipment status
		var ordersToUpdate = []*ent.Order{}
		ordersToUpdate, err = equipmentStatus.QueryOrder().All(ctx)
		if err != nil {
			return err
		}
		// change all order statuses to close
		for _, orderToUpdate := range ordersToUpdate {
			var orderStatusesToUpdate = []*ent.OrderStatus{}
			orderStatusesToUpdate, err = tx.OrderStatus.Query().QueryOrder().Where(order.ID(orderToUpdate.ID)).
				QueryOrderStatus().All(ctx)
			if err != nil {
				return err
			}
			for _, orderStatusToUpdate := range orderStatusesToUpdate {
				_, err = orderStatusToUpdate.Update().SetOrderStatusNameID(orderStatusClosed.ID).Save(ctx)
				if err != nil {
					return err
				}
			}
		}
	}
	// change equipment status to archived
	_, err = tx.Equipment.UpdateOneID(id).SetCurrentStatus(equipmentStatusArchived).Save(ctx)
	return err
}

func (r *equipmentRepository) EquipmentsByFilterTotal(ctx context.Context, filter models.EquipmentFilter) (int, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return 0, err
	}

	var filterReceiptDate string
	if filter.ReceiptDate != 0 {
		filterReceiptDate = time.Unix(filter.ReceiptDate, 0).Format(utils.TimeFormat)
	}

	total, err := tx.Equipment.Query().
		Where(
			equipment.HasCategoryWith(OptionalIntCategory(filter.Category, category.FieldID)),
			equipment.HasSubcategoryWith(OptionalIntSubcategory(filter.Subcategory, subcategory.FieldID)),
			equipment.HasCurrentStatusWith(OptionalIntStatus(filter.Status, equipmentstatusname.FieldID)),
			equipment.NameContains(filter.NameSubstring),
			OptionalStringEquipment(filter.Name, equipment.FieldName),
			OptionalStringEquipment(filter.Description, equipment.FieldDescription),
			OptionalStringEquipment(filter.TermsOfUse, equipment.FieldTermsOfUse),
			OptionalIntEquipment(filter.CompensationCost, equipment.FieldCompensationCost),
			OptionalIntEquipment(filter.InventoryNumber, equipment.FieldInventoryNumber),
			OptionalStringEquipment(filter.Supplier, equipment.FieldSupplier),
			OptionalStringEquipment(
				filterReceiptDate,
				equipment.FieldReceiptDate,
			),
			OptionalStringEquipment(filter.Title, equipment.FieldTitle),
			OptionalBoolEquipment(filter.TechnicalIssues, equipment.FieldTechIssue),
			OptionalStringEquipment(filter.Condition, equipment.FieldCondition),
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

	eqReceiptDate := time.Unix(*eq.ReceiptDate, 0).Format(utils.TimeFormat)

	if *eq.ReceiptDate != 0 {
		edit.SetReceiptDate(eqReceiptDate)
	}

	if *eq.Category != 0 {
		edit.SetCategory(&ent.Category{ID: int(*eq.Category)})
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

func (r *equipmentRepository) BlockEquipment(
	ctx context.Context, id int, startDate, endDate time.Time, userID int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}

	eqToBlock, err := tx.Equipment.Get(ctx, id)
	if err != nil {
		return err
	}

	// Get last EqupmentStatus for current Equipment
	currentEqStatus, err := tx.EquipmentStatus.
		Query().
		QueryEquipments().
		Where(equipment.IDEQ(id)).
		QueryEquipmentStatus().
		WithEquipmentStatusName().
		Order(ent.Asc(equipmentstatus.FieldEndDate)).
		First(ctx)
	if err != nil {
		return err
	}

	// Get EquipmentStatusName from DB
	eqStatusNotAvailable, err := tx.EquipmentStatusName.
		Query().
		Where(equipmentstatusname.Name(domain.EquipmentStatusNotAvailable)).
		Only(ctx)
	if err != nil {
		return err
	}

	// Set a new EquipmentStatusName for current Equipment
	_, err = eqToBlock.Update().SetCurrentStatus(eqStatusNotAvailable).Save(ctx)
	if err != nil {
		return err
	}

	// Check if current EqupmentStatus has specific EquipmentStatusName.
	// if the record exist update it, otherwise create a new EquipmentStatus.
	if currentEqStatus != nil && currentEqStatus.Edges.EquipmentStatusName.Name == domain.EquipmentStatusNotAvailable {
		currentEqStatus.
			Update().
			SetStartDate(startDate).
			SetEndDate(endDate).
			Save(ctx)
	} else {
		_, err = tx.EquipmentStatus.Create().
			SetCreatedAt(time.Now()).
			SetEndDate(endDate).
			SetStartDate(startDate).
			SetEquipments(eqToBlock).
			SetEquipmentStatusName(eqStatusNotAvailable).
			SetUpdatedAt(time.Now()).
			Save(ctx)
		if err != nil {
			return err
		}
	}

	// Get OrderStatusName form DB
	orStatusBlocked, err := tx.OrderStatusName.
		Query().
		Where(orderstatusname.Status(domain.OrderStatusBlocked)).
		Only(ctx)
	if err != nil {
		return err
	}

	// Find all Orders which have OrderStatusName booked and start from startDate and later
	orderIDs, err := eqToBlock.QueryOrder().
		Where(order.RentStartGTE(startDate), order.RentStartLTE(endDate)). // rentStart must be in range of startDate..endDate
		Where(order.RentEndGTE(startDate)).                                // rentEnd must be equal or greater than startDate
		Where(order.HasCurrentStatusWith(orderstatusname.StatusIn(
			domain.OrderStatusPrepared,
			domain.OrderStatusApproved))).
		IDs(ctx)
	if err != nil {
		return err
	}

	// Do action only if we found some Orders according to DB request above.
	if len(orderIDs) > 0 {
		// Set a new OrderStatusName for these Orders
		_, err = tx.Order.Update().Where(order.IDIn(orderIDs...)).SetCurrentStatus(orStatusBlocked).Save(ctx)
		if err != nil {
			return err
		}

		// Create new OrderStatuses for orders
		oss := make([]*ent.OrderStatusCreate, len(orderIDs))
		for i, order := range orderIDs {
			oss[i] = tx.OrderStatus.Create().
				SetComment("").
				SetCurrentDate(time.Now()).
				SetOrderID(order).
				SetUsersID(userID).
				SetOrderStatusName(orStatusBlocked)
		}
		_, err = tx.OrderStatus.CreateBulk(oss...).Save(ctx)
		if err != nil {
			return err
		}
	}
	return err
}

func (r *equipmentRepository) UnblockEquipment(ctx context.Context, id int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}

	eqToUnblock, err := tx.Equipment.Get(ctx, id)
	if err != nil {
		return err
	}

	// Get EquipmentStatusNames form DB
	eqStatusNotAvailable, err := tx.EquipmentStatusName.
		Query().
		Where(equipmentstatusname.Name(domain.EquipmentStatusNotAvailable)).
		Only(ctx)
	if err != nil {
		return err
	}

	eqStatusAvailable, err := tx.EquipmentStatusName.
		Query().
		Where(equipmentstatusname.Name(domain.EquipmentStatusAvailable)).
		Only(ctx)
	if err != nil {
		return err
	}

	// Set a new EquipmentStatusName for current Equipment
	_, err = eqToUnblock.Update().SetCurrentStatus(eqStatusAvailable).Save(ctx)
	if err != nil {
		return err
	}

	// Get last EqupmentStatus for Equipment according to some criteria
	equipmentStatus, err := tx.EquipmentStatus.
		Query().
		Where(equipmentstatus.HasEquipmentsWith(equipment.ID(eqToUnblock.ID))).
		Where(equipmentstatus.HasEquipmentStatusNameWith(equipmentstatusname.ID(eqStatusNotAvailable.ID))).
		Order(ent.Asc(equipmentstatus.FieldEndDate)).
		First(ctx)
	if err != nil {
		return err
	}

	if equipmentStatus != nil {
		_, err = equipmentStatus.
			Update().
			SetEquipmentStatusName(eqStatusAvailable).
			SetEndDate(truncateHours(&equipmentStatus.EndDate)).
			Save(ctx)
		if err != nil {
			return err
		}

	}
	return err
}

// Set EquipmentStatus EndDate back in time to give access for the unblocked equipment again
func truncateHours(date *time.Time) time.Time {
	extraHour := 1
	hours := date.Hour() + extraHour
	return date.Add(time.Duration(-hours) * time.Hour)
}
