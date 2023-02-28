package repositories

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
)

type UserSuite struct {
	suite.Suite
	ctx    context.Context
	client *ent.Client
	users  map[int]*ent.User
}

func TestUserSuite(t *testing.T) {
	s := new(UserSuite)
	suite.Run(t, s)
}

func (s *UserSuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:users?mode=memory&cache=shared&_fk=1")
	s.client = client

	s.users = make(map[int]*ent.User)
	s.users[1] = &ent.User{
		Login:    "user_1",
		Email:    "user_1@mail.com",
		Password: "password",
		Name:     "user1",
	}
	s.users[2] = &ent.User{
		Login:    "user_2",
		Email:    "user_2@mail.com",
		Password: "password",
		Name:     "user2",
	}
	s.users[3] = &ent.User{
		Login:    "user_3",
		Email:    "user_3@mail.com",
		Password: "password",
		Name:     "user3",
	}
	s.users[4] = &ent.User{
		Login:    "user_4",
		Email:    "user_4@mail.com",
		Password: "password",
		Name:     "user4",
	}
	s.users[5] = &ent.User{
		Login:    "user_5",
		Email:    "user_5@mail.com",
		Password: "password",
		Name:     "user5",
	}
	s.users[6] = &ent.User{
		Login:    "user_6",
		Email:    "user_6@mail.com",
		Password: "password",
		Name:     "user6",
	}

	_, err := s.client.User.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i, value := range s.users {
		user, errCreate := s.client.User.Create().
			SetName(value.Name).SetLogin(value.Login).SetPassword(value.Password).SetEmail(value.Email).
			Save(s.ctx)
		if errCreate != nil {
			t.Fatal(errCreate)
		}
		s.users[i].ID = user.ID
	}
}

func (s *UserSuite) TearDownSuite() {
	s.client.Close()
}

func (s *UserSuite) TestUserRepository_UsersListTotal() {
	t := s.T()
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	totalUsers, err := repository.UsersListTotal(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.users), totalUsers)
}

func (s *UserSuite) TestUserRepository_UserList_EmptyOrderBy() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := ""
	orderColumn := user.FieldID
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, users)
}

func (s *UserSuite) TestUserRepository_UserList_EmptyOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := ""
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, users)
}

func (s *UserSuite) TestUserRepository_UserList_WrongOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := user.FieldIsBlocked
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Nil(t, users)
}

func (s *UserSuite) TestUserRepository_UserList_OrderByIDDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := user.FieldID
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.users), len(users))
	prevUserID := math.MaxInt
	for _, value := range users {
		assert.True(t, mapContainsUser(t, value, s.users))
		assert.LessOrEqual(t, value.ID, prevUserID)
		prevUserID = value.ID
	}
}

func (s *UserSuite) TestUserRepository_UserList_OrderByNameDesc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.DescOrder
	orderColumn := user.FieldName
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.users), len(users))
	prevUserName := "zzzzzzzzzzzzzzzzzzzzz"
	for _, value := range users {
		assert.True(t, mapContainsUser(t, value, s.users))
		assert.LessOrEqual(t, value.Name, prevUserName)
		prevUserName = value.Name
	}
}

func (s *UserSuite) TestUserRepository_UserList_OrderByIDAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := user.FieldID
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.users), len(users))
	prevUserID := 0
	for _, value := range users {
		assert.True(t, mapContainsUser(t, value, s.users))
		assert.GreaterOrEqual(t, value.ID, prevUserID)
		prevUserID = value.ID
	}
}

func (s *UserSuite) TestUserRepository_UserList_OrderByNameAsc() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := user.FieldName
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.users), len(users))
	prevUserName := ""
	for _, value := range users {
		assert.True(t, mapContainsUser(t, value, s.users))
		assert.GreaterOrEqual(t, value.Name, prevUserName)
		prevUserName = value.Name
	}
}

func (s *UserSuite) TestUserRepository_UserList_Limit() {
	t := s.T()
	limit := 5
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := user.FieldID
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, limit, len(users))
}

func (s *UserSuite) TestUserRepository_UserList_Offset() {
	t := s.T()
	limit := 0
	offset := 5
	orderBy := utils.AscOrder
	orderColumn := user.FieldID
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	assert.NoError(t, tx.Commit())
	assert.Equal(t, len(s.users)-offset, len(users))
}

func (s *UserSuite) TestUserRepository_ChangePasswordByLogin() {
	t := s.T()
	repository := NewUserRepository()
	ctx := s.ctx
	login := s.users[1].Login
	assert.NotEmpty(t, login)
	newPassword := "password1"
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = repository.ChangePasswordByLogin(ctx, login, newPassword)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())

	tx, err = s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	user, err := repository.UserByLogin(ctx, login)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword))
	assert.NoError(t, err)
}

func mapContainsUser(t *testing.T, eq *ent.User, m map[int]*ent.User) bool {
	t.Helper()
	for _, v := range m {
		if eq.Name == v.Name && eq.ID == v.ID && eq.Login == v.Login && eq.Email == v.Email {
			return true
		}
	}
	return false
}
