package user

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
)

func TestIntegration_Logout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	loginUser, err := utils.LoginUser(ctx, client, l, p)
	require.NoError(t, err)

	t.Run("login user passed", func(t *testing.T) {
		params := users.NewLogoutParams()
		params.SetRefreshToken(&models.RefreshToken{
			RefreshToken: loginUser.GetPayload().RefreshToken,
		})
		params.SetContext(ctx)
		params.SetHTTPClient(http.DefaultClient)

		login, err := client.Users.Logout(params)
		require.NoError(t, err)
		require.NotNil(t, login.GetPayload())

		// try to use same token after logout
		refreshParams := users.NewRefreshParamsWithContext(ctx)
		refreshParams.SetHTTPClient(http.DefaultClient)
		refreshParams.SetRefreshToken(&models.RefreshToken{
			RefreshToken: loginUser.GetPayload().RefreshToken,
		})

		_, err = client.Users.Refresh(refreshParams)
		require.Error(t, err)
	})

	t.Run("logout with dummy data", func(t *testing.T) {
		params := users.NewLogoutParams()
		dummyToken := "dummy"
		params.SetRefreshToken(&models.RefreshToken{
			RefreshToken: &dummyToken,
		})
		params.SetContext(ctx)
		params.SetHTTPClient(http.DefaultClient)

		_, err = client.Users.Logout(params)
		require.Error(t, err)
	})

}
