package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type roleRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	roles      map[string]string
	client     *ent.Client
	repository domain.RoleRepository
}

func TestRoleSuite(t *testing.T) {
	suite.Run(t, new(roleRepositoryTestSuite))
}

func (s *roleRepositoryTestSuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:role?mode=memory&cache=shared&_fk=1")
	s.client = client
	s.repository = NewRoleRepository()

	_, err := s.client.Role.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	// add some roles.
	s.roles = make(map[string]string)
	s.roles["admin"] = "admin_slug"
	s.roles["user"] = "user_slug"
	for role, slug := range s.roles {
		_, err := s.client.Role.Create().SetName(role).SetSlug(slug).Save(s.ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func (s *roleRepositoryTestSuite) TearDownSuite() {
	s.client.Close()
}

func (s *roleRepositoryTestSuite) TestRoleRepository_GetRoles() {
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	roles, err := s.repository.GetRoles(ctx)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.roles), len(roles))
	for _, role := range roles {
		assert.Contains(t, s.roles, role.Name)
	}
}
