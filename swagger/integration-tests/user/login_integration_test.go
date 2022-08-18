package user

import (
	"context"
	"net/http"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/integration-tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Login(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	user, err := utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	t.Run("login user passed", func(t *testing.T) {
		info := &models.LoginInfo{
			Login:    user.Login,
			Password: &p,
		}
		params := users.NewLoginParams()
		params.SetLogin(info)
		params.SetContext(ctx)
		params.SetHTTPClient(http.DefaultClient)

		login, err := client.Users.Login(params)
		require.NoError(t, err)

		require.NotNil(t, login.GetPayload())
	})

	t.Run("login user failed", func(t *testing.T) {
		testLogin := utils.LoginNotExist
		info := &models.LoginInfo{
			Login:    &testLogin,
			Password: &p,
		}
		params := users.NewLoginParams()
		params.SetLogin(info)
		params.SetContext(ctx)
		params.SetHTTPClient(http.DefaultClient)

		_, err := client.Users.Login(params)
		require.Error(t, err)

		gotErr, ok := err.(*users.LoginUnauthorized)
		require.Equal(t, ok, true)

		wantErr := users.NewLoginUnauthorized()
		wantErr.Payload = "Invalid login or password"

		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("login user failed with password", func(t *testing.T) {
		testPassword := utils.PasswordNotExist
		info := &models.LoginInfo{
			Login:    &l,
			Password: &testPassword,
		}
		params := users.NewLoginParams()
		params.SetLogin(info)
		params.SetContext(ctx)
		params.SetHTTPClient(http.DefaultClient)

		_, err := client.Users.Login(params)
		require.Error(t, err)

		gotErr, ok := err.(*users.LoginUnauthorized)
		require.Equal(t, ok, true)

		wantErr := users.NewLoginUnauthorized()
		wantErr.Payload = "Invalid login or password"

		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("failed if no data provided", func(t *testing.T) {
		info := &models.LoginInfo{}
		params := users.NewLoginParams()
		params.SetLogin(info)
		params.SetContext(ctx)
		params.SetHTTPClient(http.DefaultClient)

		_, err := client.Users.Login(params)
		require.Error(t, err)

		gotErr, ok := err.(*users.LoginDefault)
		require.Equal(t, ok, true)

		wantErr := users.NewLoginDefault(422)
		wantErr.Payload = &models.Error{Data: nil}

		assert.Equal(t, wantErr, gotErr)
	})
}
