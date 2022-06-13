package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
)

type ActiveAreasSuite struct {
	suite.Suite
	ctx         context.Context
	client      *ent.Client
	repository  ActiveAreaRepository
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

	s.activeAreas = make(map[int]string)
	s.activeAreas[1] = "area 1"
	s.activeAreas[2] = "area 2"

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

func (s *ActiveAreasSuite) TestActiveAreaRepository_AllActiveAreas() {
	t := s.T()
	repository := NewActiveAreaRepository(s.client)
	activeAreas, err := repository.AllActiveAreas(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(s.activeAreas), len(activeAreas))
	assert.Equal(t, s.activeAreas[1], activeAreas[0].Name)
	assert.Equal(t, s.activeAreas[2], activeAreas[1].Name)
}
