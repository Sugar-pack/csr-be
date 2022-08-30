package user

import (
	"context"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/integration-tests/common"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Refresh(t *testing.T) {
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

	t.Run("refresh token passed", func(t *testing.T) {
		refreshToken := &models.RefreshToken{
			RefreshToken: loginUser.GetPayload().RefreshToken,
		}
		params := users.NewRefreshParamsWithContext(ctx)
		params.SetRefreshToken(refreshToken)

		refresh, err := client.Users.Refresh(params)
		require.NoError(t, err)

		newToken := refresh.GetPayload().AccessToken
		assert.NotNil(t, newToken)
	})

	t.Run("access token also valid, passed", func(t *testing.T) {
		refreshToken := &models.RefreshToken{
			RefreshToken: loginUser.GetPayload().AccessToken,
		}

		params := users.NewRefreshParamsWithContext(ctx)
		params.SetRefreshToken(refreshToken)

		refresh, gotErr := client.Users.Refresh(params)
		require.NoError(t, gotErr)

		newToken := refresh.GetPayload().AccessToken
		assert.NotNil(t, newToken)
	})

	t.Run("refresh token failed: invalid token", func(t *testing.T) {
		invalidToken := "invalid"
		refreshToken := &models.RefreshToken{
			RefreshToken: &invalidToken,
		}
		params := users.NewRefreshParamsWithContext(ctx)
		params.SetRefreshToken(refreshToken)

		_, gotErr := client.Users.Refresh(params)
		require.Error(t, gotErr)

		wantErr := users.NewRefreshDefault(400)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{
			Message: "token invalid",
		}}
		assert.Equal(t, wantErr, gotErr)
	})
}
