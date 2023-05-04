package repositories

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
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
			SetName(value.Name).
			SetLogin(value.Login).
			SetPassword(value.Password).
			SetEmail(value.Email).
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
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	totalUsers, err := repository.UsersListTotal(ctx)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.users), totalUsers)
}

func (s *UserSuite) TestUserRepository_UsersListTotal_NoDeletedUsersInList() {
	// Setup test transaction and repo
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	repo := NewUserRepository()

	// Insert a deleted user
	_, err = tx.User.Create().
		SetLogin("test_deleted").
		SetEmail("test_deleted").
		SetPassword("test_deleted").
		SetName("test_deleted").
		SetIsDeleted(true).
		Save(ctx)
	require.NoError(t, err)

	// Check
	totalUsers, err := repo.UsersListTotal(ctx)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, err)
	require.Equal(t, len(s.users), totalUsers)

	require.NoError(s.T(), tx.Rollback())
}

func (s *UserSuite) TestUserRepository_UsersList_NoDeletedUsersInList() {
	// Setup test transaction and repo
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	repo := NewUserRepository()

	// Insert a deleted user
	_, err = tx.User.Create().
		SetLogin("test_deleted").
		SetEmail("test_deleted").
		SetPassword("test_deleted").
		SetName("test_deleted").
		SetIsDeleted(true).
		Save(ctx)
	require.NoError(t, err)

	// Check
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := user.FieldID
	users, err := repo.UserList(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, err)
	require.Equal(t, len(s.users), len(users))

	require.NoError(s.T(), tx.Rollback())
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
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, users)
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
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, users)
}

func (s *UserSuite) TestUserRepository_UserList_WrongOrderColumn() {
	t := s.T()
	limit := math.MaxInt
	offset := 0
	orderBy := utils.AscOrder
	orderColumn := user.FieldIsReadonly
	repository := NewUserRepository()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Nil(t, users)
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
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.users), len(users))
	prevUserID := math.MaxInt
	for _, value := range users {
		require.True(t, mapContainsUser(t, value, s.users))
		require.LessOrEqual(t, value.ID, prevUserID)
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
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.users), len(users))
	prevUserName := "zzzzzzzzzzzzzzzzzzzzz"
	for _, value := range users {
		require.True(t, mapContainsUser(t, value, s.users))
		require.LessOrEqual(t, value.Name, prevUserName)
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
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.users), len(users))
	prevUserID := 0
	for _, value := range users {
		require.True(t, mapContainsUser(t, value, s.users))
		require.GreaterOrEqual(t, value.ID, prevUserID)
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
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.users), len(users))
	prevUserName := ""
	for _, value := range users {
		require.True(t, mapContainsUser(t, value, s.users))
		require.GreaterOrEqual(t, value.Name, prevUserName)
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
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, limit, len(users))
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
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	users, err := repository.UserList(ctx, limit, offset, orderBy, orderColumn)
	if err != nil {
		t.Fatal(err)
	}
	require.NoError(t, tx.Commit())
	require.Equal(t, len(s.users)-offset, len(users))
}

func (s *UserSuite) TestUserRepository_ChangePasswordByLogin() {
	// Setup test transaction and repo
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	repo := NewUserRepository()

	// Insert a test user
	user, err := tx.User.Create().
		SetLogin("test_change_password").
		SetEmail("test_change_password").
		SetPassword("test_change_password").
		SetName("test_change_password").
		Save(ctx)
	require.NoError(t, err)

	// Call ChangePasswordByLogin
	newPassword := "password1"
	err = repo.ChangePasswordByLogin(ctx, user.Login, newPassword)
	require.NoError(t, err)

	// Check updated
	updatedUser, err := tx.User.Get(ctx, user.ID)
	require.NoError(t, err)
	err = bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte(newPassword))
	require.NoError(t, err)

	require.NoError(s.T(), tx.Rollback())
}

func (s *UserSuite) TestUserRepository_SetIsReadonly() {
	// Setup test transaction and repo
	t := s.T()
	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	repo := NewUserRepository()

	// Insert a test user
	user, err := tx.User.Create().
		SetLogin("test_readonly").
		SetEmail("test_readonly").
		SetPassword("test_readonly").
		SetName("test_readonly").
		SetIsReadonly(false).
		Save(ctx)
	require.NoError(t, err)

	// Call SetIsReadonly
	err = repo.SetIsReadonly(ctx, user.ID, true)
	require.NoError(t, err)

	// Check updated
	updatedUser, err := tx.User.Get(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, updatedUser.IsReadonly)

	// Call SetIsReadonly with an invalid ID
	err = repo.SetIsReadonly(ctx, 9999, true)
	require.Error(t, err)

	require.NoError(s.T(), tx.Rollback())
}

func (s *UserSuite) TestUserRepository_DeleteUser_OK() {
	t := s.T()
	repository := NewUserRepository()
	ctx := s.ctx
	login := s.users[1].Login
	require.NotEmpty(t, login)

	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	user, err := repository.UserByLogin(ctx, login)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	assert.Equal(t, user.IsDeleted, false)

	tx, err = s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = repository.Delete(ctx, user.ID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	tx, err = s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	updatedUser, err := repository.UserByLogin(ctx, login)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	assert.Equal(t, updatedUser.IsDeleted, true)
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
