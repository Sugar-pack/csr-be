package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"

	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

type blockerTestSuite struct {
	suite.Suite
	ctx        context.Context
	repository BlockerRepository
	client     *ent.Client
	user       *ent.User
}

func TestBlockerSuite(t *testing.T) {
	suite.Run(t, new(blockerTestSuite))
}

func (s *blockerTestSuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:blocker?mode=memory&cache=shared&_fk=1")
	s.client = client

	_, err := s.client.User.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	user, err := s.client.User.Create().SetLogin("admin").SetName("user"). // create user
										SetPassword("admin").SetEmail("test@example.com").Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	s.user = user

	s.repository = NewBlockerRepository(s.client)
}

func (s *blockerTestSuite) TearDownSuite() {
	s.client.Close()
}

func (s *blockerTestSuite) TestBlockerRepository_SetIsBlockedUser_SetTrue() {
	t := s.T()
	err := s.repository.SetIsBlockedUser(s.ctx, s.user.ID, true)
	assert.NoError(t, err)
	updatedUser, err := s.client.User.Get(s.ctx, s.user.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, updatedUser.IsBlocked)
}

func (s *blockerTestSuite) TestBlockerRepository_SetIsBlockedUser_SetFalse() {
	t := s.T()
	err := s.repository.SetIsBlockedUser(s.ctx, s.user.ID, false)
	assert.NoError(t, err)
	updatedUser, err := s.client.User.Get(s.ctx, s.user.ID)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, updatedUser.IsBlocked)
}

func (s *blockerTestSuite) TestBlockerRepository_SetIsBlockedUser_NoUser() {
	t := s.T()
	err := s.client.User.DeleteOneID(s.user.ID).Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = s.repository.SetIsBlockedUser(s.ctx, s.user.ID, false)
	assert.Error(t, err)
}
