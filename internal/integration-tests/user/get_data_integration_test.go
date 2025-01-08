package user

import (
	"context"
	"net/http"
	"testing"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testLogin    string
	testPassword string
	testUserID   int64
)

func TestIntegration_GetCurrentUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, id := utils.CreateLoginPassword(t, 1)

	testLogin = l
	testPassword = p
	testUserID = id

	loginUser, err := utils.LoginUser(ctx, client, l, p)

	t.Run("get user data passed", func(t *testing.T) {
		params := users.NewGetCurrentUserParamsWithContext(ctx)
		authInfo := utils.AuthInfoFunc(loginUser.GetPayload().AccessToken)

		currentUser, err := client.Users.GetCurrentUser(params, authInfo)
		require.NoError(t, err)

		assert.Equal(t, *currentUser.Payload.Login, l)
		assert.Equal(t, *currentUser.Payload.ID, testUserID)
	})

	t.Run("get current user data failed: token contains an invalid number of segments", func(t *testing.T) {
		params := users.NewGetCurrentUserParamsWithContext(ctx)
		dummyToken := utils.TokenNotExist
		authInfo := utils.AuthInfoFunc(&dummyToken)

		_, err = client.Users.GetCurrentUser(params, authInfo)

		errExp := users.NewGetCurrentUserDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("get current user data failed: no authorization", func(t *testing.T) {
		params := users.NewGetCurrentUserParamsWithContext(ctx)
		authInfo := utils.AuthInfoFunc(nil)

		_, err = client.Users.GetCurrentUser(params, authInfo)
		assert.Error(t, err)

		errExp := users.NewGetCurrentUserDefault(401)
		msgExp := "unauthenticated for invalid credentials"
		codeExp := int32(http.StatusUnauthorized)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}
		assert.Equal(t, errExp, err)
	})
}

func TestIntegration_GetAllUsers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	loginUser1 := utils.AdminUserLogin(t)
	// todo: get rid of access interface{} in GetAllUsersHandlerFunc (not used)
	t.Run("get all users passed", func(t *testing.T) {
		params := users.NewGetAllUsersParamsWithContext(ctx)
		authInfo := utils.AuthInfoFunc(loginUser1.GetPayload().AccessToken)

		gotUsers, err := client.Users.GetAllUsers(params, authInfo)
		assert.NoError(t, err)

		assert.NotZero(t, gotUsers.GetPayload())
	})

	t.Run("get all users failed: no authorization", func(t *testing.T) {
		params := users.NewGetAllUsersParamsWithContext(ctx)
		authInfo := utils.AuthInfoFunc(nil)

		_, err := client.Users.GetAllUsers(params, authInfo)
		assert.Error(t, err)

		errExp := users.NewGetAllUsersDefault(401)
		msgExp := "unauthenticated for invalid credentials"
		codeExp := int32(http.StatusUnauthorized)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("get all user data failed: token contains an invalid number of segments", func(t *testing.T) {
		params := users.NewGetAllUsersParamsWithContext(ctx)
		dummyToken := utils.TokenNotExist
		authInfo := utils.AuthInfoFunc(&dummyToken)

		_, err := client.Users.GetAllUsers(params, authInfo)

		errExp := users.NewGetAllUsersDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, errExp, err)
	})
}

func TestIntegration_GetUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()
	time.Sleep(time.Second)

	loginUser, err := utils.LoginUser(ctx, client, testLogin, testPassword)
	require.NoError(t, err)

	params := users.NewGetUserParamsWithContext(ctx)
	params.UserID = testUserID
	authInfo := utils.AuthInfoFunc(loginUser.GetPayload().AccessToken)

	user, err := client.Users.GetUser(params, authInfo)
	require.NoError(t, err)

	gotID := user.GetPayload().ID
	assert.Equal(t, &testUserID, gotID)
}
