package repositories

import (
	"context"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type kindRepositorySuite struct {
	suite.Suite
	ctx        context.Context
	client     *ent.Client
	repository KindRepository
	kinds      []*ent.Kind
}

func TestKindSuite(t *testing.T) {
	suite.Run(t, new(kindRepositorySuite))
}

func (s *kindRepositorySuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:kind?mode=memory&cache=shared&_fk=1")
	s.client = client
	s.repository = NewKindRepository(client)

	s.kinds = []*ent.Kind{
		{
			Name:                "kind 1",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "kind 2",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "kind 3",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "kind 4",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
		{
			Name:                "kind 5",
			MaxReservationTime:  int64(10),
			MaxReservationUnits: int64(2),
		},
	}
	_, err := s.client.Kind.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range s.kinds {
		kind, errCreate := s.client.Kind.Create().
			SetName(value.Name).SetMaxReservationTime(value.MaxReservationTime).
			SetMaxReservationUnits(value.MaxReservationUnits).
			Save(s.ctx)
		if errCreate != nil {
			t.Fatal(errCreate)
		}
		s.kinds[i].ID = kind.ID
	}
}

func (s *kindRepositorySuite) TearDownSuite() {
	s.client.Close()
}

func (s *kindRepositorySuite) TestKindRepository_CreateKind() {
	t := s.T()
	name := "kind"
	maxReservationTime := int64(10)
	maxReservationUnits := int64(1)
	newKind := models.CreateNewKind{
		Name:                &name,
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnits,
	}
	createdKind, err := s.repository.CreateKind(s.ctx, newKind)
	assert.NoError(t, err)
	assert.Equal(t, name, createdKind.Name)
	assert.Equal(t, maxReservationTime, createdKind.MaxReservationTime)
	assert.Equal(t, maxReservationUnits, createdKind.MaxReservationUnits)

	_, err = s.client.Kind.Delete().Where(kind.IDEQ(createdKind.ID)).Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *kindRepositorySuite) TestKindRepository_AllKindsTotal() {
	t := s.T()
	total, err := s.repository.AllKindsTotal(s.ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(s.kinds), total)
}
func (s *kindRepositorySuite) TestKindRepository_AllKind_EmptyOrderBy() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := ""
	orderColumn := kind.FieldID
	kinds, err := s.repository.AllKinds(s.ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.Nil(t, kinds)
}

func (s *kindRepositorySuite) TestKindRepository_AllKind_EmptyOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := ""
	kinds, err := s.repository.AllKinds(s.ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.Nil(t, kinds)
}

func (s *kindRepositorySuite) TestKindRepository_AllKind_WrongOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := kind.FieldMaxReservationTime
	kinds, err := s.repository.AllKinds(s.ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.Nil(t, kinds)
}

func (s *kindRepositorySuite) TestKindRepository_AllKind_OrderByIDDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := kind.FieldID
	kinds, err := s.repository.AllKinds(s.ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, len(s.kinds), len(kinds))
	prevKindID := math.MaxInt
	for _, value := range kinds {
		assert.True(t, containsKind(t, value, s.kinds))
		assert.LessOrEqual(t, value.ID, prevKindID)
		prevKindID = value.ID
	}
}

func (s *kindRepositorySuite) TestKindRepository_AllKind_OrderByNameDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := kind.FieldName
	kinds, err := s.repository.AllKinds(s.ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, len(s.kinds), len(kinds))
	prevKindName := "zzzzzzzzzzzzzzzzzzzzz"
	for _, value := range kinds {
		assert.True(t, containsKind(t, value, s.kinds))
		assert.LessOrEqual(t, value.Name, prevKindName)
		prevKindName = value.Name
	}
}

func (s *kindRepositorySuite) TestKindRepository_AllKind_OrderByIDAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := kind.FieldID
	kinds, err := s.repository.AllKinds(s.ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, len(s.kinds), len(kinds))
	prevKindID := 0
	for _, value := range kinds {
		assert.True(t, containsKind(t, value, s.kinds))
		assert.GreaterOrEqual(t, value.ID, prevKindID)
		prevKindID = value.ID
	}
}

func (s *kindRepositorySuite) TestKindRepository_AllKind_OrderByNameAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := kind.FieldName
	kinds, err := s.repository.AllKinds(s.ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.Equal(t, len(s.kinds), len(kinds))
	prevKindName := ""
	for _, value := range kinds {
		assert.True(t, containsKind(t, value, s.kinds))
		assert.GreaterOrEqual(t, value.Name, prevKindName)
		prevKindName = value.Name
	}
}

func (s *kindRepositorySuite) TestKindRepository_AllKind_Limit() {
	t := s.T()
	limit := 5
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := kind.FieldID
	kinds, err := s.repository.AllKinds(s.ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, limit, len(kinds))
}

func (s *kindRepositorySuite) TestKindRepository_AllKind_Offset() {
	t := s.T()
	limit := 0
	offset := 5
	orderBy := utils.AscOrder
	orderColumn := kind.FieldID
	kinds, err := s.repository.AllKinds(s.ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(s.kinds)-offset, len(kinds))
}

func (s *kindRepositorySuite) TestKindRepository_KindByID() {
	t := s.T()
	kind, err := s.repository.KindByID(s.ctx, s.kinds[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, s.kinds[0].Name, kind.Name)
	assert.Equal(t, s.kinds[0].MaxReservationTime, kind.MaxReservationTime)
	assert.Equal(t, s.kinds[0].MaxReservationUnits, kind.MaxReservationUnits)

	_, err = s.client.Kind.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *kindRepositorySuite) TestKindRepository_DeleteKindByID() {
	t := s.T()
	name := "kind"
	maxReservationTime := int64(10)
	maxReservationUnits := int64(1)
	createdKind, err := s.client.Kind.Create().SetName(name).
		SetMaxReservationTime(maxReservationTime).
		SetMaxReservationUnits(maxReservationUnits).
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = s.repository.DeleteKindByID(s.ctx, createdKind.ID)
	assert.NoError(t, err)
}

func (s *kindRepositorySuite) TestKindRepository_UpdateKind() {
	t := s.T()
	update := models.PatchKind{
		Name: "kind 0",
	}
	kind, err := s.repository.UpdateKind(s.ctx, s.kinds[0].ID, update)
	assert.NoError(t, err)
	assert.Equal(t, update.Name, kind.Name)
	assert.Equal(t, s.kinds[0].MaxReservationTime, kind.MaxReservationTime)
	assert.Equal(t, s.kinds[0].MaxReservationUnits, kind.MaxReservationUnits)
}

func containsKind(t *testing.T, eq *ent.Kind, list []*ent.Kind) bool {
	t.Helper()
	for _, v := range list {
		if eq.Name == v.Name && eq.ID == v.ID && eq.MaxReservationUnits == v.MaxReservationUnits &&
			eq.MaxReservationTime == v.MaxReservationTime {
			return true
		}
	}
	return false
}
