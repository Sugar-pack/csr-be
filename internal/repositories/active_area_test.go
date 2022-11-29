package repositories

import (
	"context"
	"math"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/activearea"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type ActiveAreasSuite struct {
	suite.Suite
	ctx         context.Context
	client      *ent.Client
	repository  domain.ActiveAreaRepository
	activeAreas map[int]string
}

func TestActiveAreaSuite(t *testing.T) {
	s := new(ActiveAreasSuite)
	suite.Run(t, s)
}

func (s *ActiveAreasSuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:activeareas?mode=memory&cache=shared&_fk=1")
	s.client = client
	s.repository = NewActiveAreaRepository()

	s.activeAreas = make(map[int]string)
	s.activeAreas[1] = "area 1"
	s.activeAreas[2] = "area 2"
	s.activeAreas[3] = "area 3"
	s.activeAreas[4] = "area 4"
	s.activeAreas[5] = "area 5"
	s.activeAreas[6] = "area 6"
	s.activeAreas[7] = "area 7"
	s.activeAreas[8] = "area 8"
	s.activeAreas[9] = "area 9"

	_, err := s.client.ActiveArea.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range s.activeAreas {
		_, errCreate := s.client.ActiveArea.Create().SetName(value).Save(s.ctx)
		if errCreate != nil {
			t.Fatal(errCreate)
		}
	}
}

func (s *ActiveAreasSuite) TearDownSuite() {
	s.client.Close()
}

func (s *ActiveAreasSuite) TestActiveAreaRepository_AllActiveAreasEmptyOrderBy() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := ""
	orderColumn := activearea.FieldName
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	activeAreas, err := s.repository.AllActiveAreas(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, activeAreas)
}

func (s *ActiveAreasSuite) TestActiveAreaRepository_AllActiveAreasEmptyOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := ""
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	activeAreas, err := s.repository.AllActiveAreas(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, activeAreas)
}

func (s *ActiveAreasSuite) TestActiveAreaRepository_AllActiveAreasOrderByNameDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := activearea.FieldName
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	activeAreas, err := s.repository.AllActiveAreas(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.activeAreas), len(activeAreas))
	prevAreaName := "zzzzzzzzzzzzzzzzzzzzzzzzz"
	for _, value := range activeAreas {
		assert.True(t, mapContainsArea(t, value.Name, s.activeAreas))
		assert.GreaterOrEqual(t, prevAreaName, value.Name)
		prevAreaName = value.Name
	}
}

func (s *ActiveAreasSuite) TestActiveAreaRepository_AllActiveAreasOrderByIDDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := activearea.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	activeAreas, err := s.repository.AllActiveAreas(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.activeAreas), len(activeAreas))
	prevAreaID := math.MaxInt
	for _, value := range activeAreas {
		assert.True(t, mapContainsArea(t, value.Name, s.activeAreas))
		assert.Less(t, value.ID, prevAreaID)
		prevAreaID = value.ID
	}
}

func (s *ActiveAreasSuite) TestActiveAreaRepository_AllActiveAreasOrderByNameAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := activearea.FieldName
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	activeAreas, err := s.repository.AllActiveAreas(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.activeAreas), len(activeAreas))
	prevAreaName := ""
	for _, value := range activeAreas {
		assert.True(t, mapContainsArea(t, value.Name, s.activeAreas))
		assert.LessOrEqual(t, prevAreaName, value.Name)
		prevAreaName = value.Name
	}
}

func (s *ActiveAreasSuite) TestActiveAreaRepository_AllActiveAreasOrderByIDAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := activearea.FieldID
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	activeAreas, err := s.repository.AllActiveAreas(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.activeAreas), len(activeAreas))
	prevAreaID := 0
	for _, value := range activeAreas {
		assert.True(t, mapContainsArea(t, value.Name, s.activeAreas))
		assert.Greater(t, value.ID, prevAreaID)
		prevAreaID = value.ID
	}
}

func (s *ActiveAreasSuite) TestActiveAreaRepository_LimitActiveAreas() {
	t := s.T()
	limit := 3
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := activearea.FieldName
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	activeAreas, err := s.repository.AllActiveAreas(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.GreaterOrEqual(t, limit, len(activeAreas))
}

func (s *ActiveAreasSuite) TestActiveAreaRepository_OffsetActiveAreas() {
	t := s.T()
	limit := 6
	offset := 6
	orderBy := utils.AscOrder
	orderColumn := activearea.FieldName
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	activeAreas, err := s.repository.AllActiveAreas(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.activeAreas)-offset, len(activeAreas))
}

func (s *ActiveAreasSuite) TestActiveAreaRepository_TotalActiveAreas() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	totalAreas, err := s.repository.TotalActiveAreas(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.activeAreas), totalAreas)
}

func mapContainsArea(t *testing.T, value string, m map[int]string) bool {
	t.Helper()
	for _, v := range m {
		if value == v {
			return true
		}
	}
	return false
}
