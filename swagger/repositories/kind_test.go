package repositories

import (
	"context"
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

	_, err = s.client.Kind.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *kindRepositorySuite) TestKindRepository_AllKind() {
	t := s.T()
	name := "kind"
	maxReservationTime := int64(10)
	maxReservationUnits := int64(1)
	_, err := s.client.Kind.Create().SetName(name).
		SetMaxReservationTime(maxReservationTime).
		SetMaxReservationUnits(maxReservationUnits).
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	kinds, err := s.repository.AllKind(s.ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(kinds))
	assert.Equal(t, name, kinds[0].Name)
	assert.Equal(t, maxReservationTime, kinds[0].MaxReservationTime)
	assert.Equal(t, maxReservationUnits, kinds[0].MaxReservationUnits)

	_, err = s.client.Kind.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *kindRepositorySuite) TestKindRepository_KindByID() {
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
	kind, err := s.repository.KindByID(s.ctx, createdKind.ID)
	assert.NoError(t, err)
	assert.Equal(t, name, kind.Name)
	assert.Equal(t, maxReservationTime, kind.MaxReservationTime)
	assert.Equal(t, maxReservationUnits, kind.MaxReservationUnits)

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

	_, err = s.client.Kind.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *kindRepositorySuite) TestKindRepository_UpdateKind() {
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
	update := models.PatchKind{
		Name: "newKind",
	}
	kind, err := s.repository.UpdateKind(s.ctx, createdKind.ID, update)
	assert.NoError(t, err)
	assert.Equal(t, update.Name, kind.Name)
	assert.Equal(t, maxReservationTime, kind.MaxReservationTime)
	assert.Equal(t, maxReservationUnits, kind.MaxReservationUnits)

	_, err = s.client.Kind.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}
