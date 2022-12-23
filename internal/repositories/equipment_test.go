package repositories

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type EquipmentSuite struct {
	suite.Suite
	ctx        context.Context
	client     *ent.Client
	repository domain.EquipmentRepository
	equipments map[int]*ent.Equipment
}

func TestEquipmentSuite(t *testing.T) {
	s := new(EquipmentSuite)
	suite.Run(t, s)
}

func (s *EquipmentSuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:equipments?mode=memory&cache=shared&_fk=1")
	s.client = client
	s.repository = NewEquipmentRepository()

	statusName := "status"
	_, err := s.client.EquipmentStatusName.Delete().Exec(s.ctx) // clean up
	if err != nil {
		t.Fatal(err)
	}
	status, err := s.client.EquipmentStatusName.Create().SetName(statusName).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	categoryName := "category"
	_, err = s.client.Category.Delete().Exec(s.ctx) // clean up
	if err != nil {
		t.Fatal(err)
	}
	category, err := s.client.Category.Create().SetName(categoryName).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	photoID := "photoID"
	_, err = s.client.Photo.Delete().Exec(s.ctx) // clean up
	require.NoError(t, err)

	photo, err := s.client.Photo.Create().SetID(photoID).Save(s.ctx)
	require.NoError(t, err)

	subcategoryName := "subcategory"
	_, err = s.client.Subcategory.Delete().Exec(s.ctx) // clean up
	if err != nil {
		t.Fatal(err)
	}
	subcategory, err := s.client.Subcategory.Create().SetName(subcategoryName).SetCategory(category).Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	petSize, err := s.client.PetSize.Create().SetSize("testSize").SetName("testSizeName").Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.client.PetKind.Delete().Exec(s.ctx) // clean up
	if err != nil {
		t.Fatal(err)
	}

	petKind, err := s.client.PetKind.Create().SetName("testNamePetKind").Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	s.equipments = make(map[int]*ent.Equipment)
	s.equipments[1] = &ent.Equipment{
		Name:  "test 1",
		Title: "equipment 1",
	}
	s.equipments[2] = &ent.Equipment{
		Name:  "equipment 2",
		Title: "equipment 2",
	}
	s.equipments[3] = &ent.Equipment{
		Name:  "test 3",
		Title: "equipment 3",
	}
	s.equipments[4] = &ent.Equipment{
		Name:  "equipment 4",
		Title: "equipment 4",
	}
	s.equipments[5] = &ent.Equipment{
		Name:        "test 5",
		Title:       "equipment 5",
		TermsOfUse:  "https://site.com",
		TechIssue:   "есть",
		Supplier:    "Виталий",
		ReceiptDate: "2018",
		Description: "удовлетворительное, местами облупляется краска",
		Condition:   "WARNING: do not put on cats!",
	}

	_, err = s.client.Equipment.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range s.equipments {
		eq, errCreate := s.client.Equipment.Create().
			SetName(value.Name).
			SetTitle(value.Title).
			SetTermsOfUse(value.TermsOfUse).
			SetTechIssue(value.TechIssue).
			SetSupplier(value.Supplier).
			SetReceiptDate(value.ReceiptDate).
			SetDescription(value.Description).
			SetCondition(value.Condition).
			SetCurrentStatus(status).
			SetCategory(category).
			SetSubcategory(subcategory).
			SetPhoto(photo).
			SetPetSizeID(petSize.ID).AddPetKinds(petKind).
			Save(s.ctx)
		if errCreate != nil {
			t.Fatal(errCreate)
		}
		s.equipments[i].ID = eq.ID
	}
}

func (s *EquipmentSuite) TearDownSuite() {
	s.client.Close()
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsEmptyOrderBy() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := ""
	orderColumn := equipment.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, equipments)
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsWrongOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := ""
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.EquipmentsByFilter(ctx, models.EquipmentFilter{},
		limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, equipments)
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsOrderColumnNotExists() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := "price"
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, equipments)
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsOrderByIDDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := equipment.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.equipments), len(equipments))
	prevEquipmentID := math.MaxInt
	for _, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Less(t, value.ID, prevEquipmentID)
		prevEquipmentID = value.ID
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsOrderByNameDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := equipment.FieldName
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.equipments), len(equipments))
	prevEquipmentName := "zzzzzzzzzzzzzzzzzzzzzzzzzzz"
	for _, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Less(t, value.Name, prevEquipmentName)
		prevEquipmentName = value.Name
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsOrderByTitleDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := equipment.FieldTitle
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.equipments), len(equipments))
	prevEquipmentTitle := "zzzzzzzzzzzzzzzzzzzzzzzzzzz"
	for _, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Less(t, value.Title, prevEquipmentTitle)
		prevEquipmentTitle = value.Title
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsOrderByIDAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := equipment.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.equipments), len(equipments))
	prevEquipmentID := 0
	for _, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Greater(t, value.ID, prevEquipmentID)
		prevEquipmentID = value.ID
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsOrderByNameAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := equipment.FieldName
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.equipments), len(equipments))
	prevEquipmentName := ""
	for _, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Greater(t, value.Name, prevEquipmentName)
		prevEquipmentName = value.Name
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsOrderByTitleAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := equipment.FieldTitle
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.equipments), len(equipments))
	prevEquipmentTitle := ""
	for _, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Greater(t, value.Title, prevEquipmentTitle)
		prevEquipmentTitle = value.Title
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsLimit() {
	t := s.T()
	limit := 3
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := equipment.FieldTitle
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, 3, len(equipments))
	for i, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Equal(t, s.equipments[i+1].Name, value.Name)
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsOffset() {
	t := s.T()
	limit := 3
	offset := 3
	orderBy := utils.AscOrder
	orderColumn := equipment.FieldTitle
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.AllEquipments(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, 2, len(equipments))
	for i, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Equal(t, s.equipments[i+1+offset].Name, value.Name)
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_AllEquipmentsTotal() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	totalEquipment, err := s.repository.AllEquipmentsTotal(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.equipments), totalEquipment)
}

func (s *EquipmentSuite) TestEquipmentRepository_FindEquipmentsOrderByTitleDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := "desc"
	orderColumn := "title"
	filter := models.EquipmentFilter{NameSubstring: "test"}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.EquipmentsByFilter(ctx, filter, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, 3, len(equipments))
	prevEquipmentTitle := "zzzzzzzzzzzzzzzzzzzzzz"
	for _, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Contains(t, value.Name, filter.NameSubstring)
		assert.Less(t, value.Title, prevEquipmentTitle)
		prevEquipmentTitle = value.Title
	}
}

// TODO fix later
//func (s *EquipmentSuite) TestEquipmentRepository_FindEquipment_CaseInsensitiveString() {
//	t := s.T()
//	limit := 1
//	offset := 0
//	orderBy := "asc"
//	orderColumn := "title"
//	filter := models.EquipmentFilter{
//		Name:            "tEsT 5",
//		Title:           "EQuiPmeNT 5",
//		TermsOfUse:      "htTps://SITE.coM",
//		TechnicalIssues: "естЬ",
//		Supplier:        "виталий",
//		ReceiptDate:     "2018",
//		Description:     "удовлетворительное, МЕСТАМИ облупляется краска",
//		Condition:       "warning: do not PUT on cats!",
//	}
//	ctx := s.ctx
//	tx, err := s.client.Tx(ctx)
//	require.NoError(t, err)
//	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
//	equipments, err := s.repository.EquipmentsByFilter(ctx, filter, limit, offset, orderBy, orderColumn)
//	require.NoError(t, err)
//	require.NoError(t, tx.Commit())
//	assert.Equal(t, 1, len(equipments))
//
//	for _, value := range equipments {
//		assert.True(t, strings.EqualFold(value.Name, filter.Name))
//		assert.True(t, strings.EqualFold(value.Title, filter.Title))
//		assert.True(t, strings.EqualFold(value.Description, filter.Description))
//		assert.True(t, strings.EqualFold(value.TermsOfUse, filter.TermsOfUse))
//		assert.True(t, strings.EqualFold(value.TechIssue, filter.TechnicalIssues))
//		assert.True(t, strings.EqualFold(value.Supplier, filter.Supplier))
//		assert.True(t, strings.EqualFold(value.ReceiptDate, filter.ReceiptDate))
//		assert.True(t, strings.EqualFold(value.Condition, filter.Condition))
//	}
//}

func (s *EquipmentSuite) TestEquipmentRepository_FindEquipmentsLimit() {
	t := s.T()
	limit := 2
	offset := 0
	orderBy := "asc"
	orderColumn := "title"
	filter := models.EquipmentFilter{NameSubstring: "test"}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.EquipmentsByFilter(ctx, filter, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, 2, len(equipments))
	prevEquipmentTitle := ""
	for _, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Contains(t, value.Name, filter.NameSubstring)
		assert.Greater(t, value.Title, prevEquipmentTitle)
		prevEquipmentTitle = value.Title
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_FindEquipmentsOffset() {
	t := s.T()
	limit := 2
	offset := 2
	orderBy := "asc"
	orderColumn := "name"
	name := "test"
	filter := models.EquipmentFilter{NameSubstring: "tEsT"}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	equipments, err := s.repository.EquipmentsByFilter(ctx, filter, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, 1, len(equipments))
	for _, value := range equipments {
		assert.True(t, mapContainsEquipment(value, s.equipments))
		assert.Contains(t, value.Name, name)
	}
}

func (s *EquipmentSuite) TestEquipmentRepository_FindEquipmentsTotal() {
	t := s.T()
	filter := models.EquipmentFilter{NameSubstring: "test"}
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	totalEquipment, err := s.repository.EquipmentsByFilterTotal(ctx, filter)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, 3, totalEquipment)
}

func mapContainsEquipment(eq *ent.Equipment, m map[int]*ent.Equipment) bool {
	for _, v := range m {
		if eq.Name == v.Name && eq.Title == v.Title {
			return true
		}
	}
	return false
}
