package user

import (
	"context"
	"testing"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"

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

	var newRefreshToken *string
	t.Run("refresh token passed", func(t *testing.T) {
		refreshToken := &models.RefreshToken{
			RefreshToken: loginUser.GetPayload().RefreshToken,
		}
		params := users.NewRefreshParamsWithContext(ctx)
		params.SetRefreshToken(refreshToken)

		time.Sleep(1 * time.Second) // it's needed to avoid the same time of token creation

		refresh, err := client.Users.Refresh(params)
		require.NoError(t, err)

		newAccessToken := refresh.GetPayload().AccessToken
		assert.NotNil(t, newAccessToken)
		newRefreshToken = refresh.GetPayload().RefreshToken
		assert.NotNil(t, newRefreshToken)
		assert.NotEqual(t, *loginUser.GetPayload().RefreshToken, *newRefreshToken)
		assert.NotEqual(t, *loginUser.GetPayload().AccessToken, *newAccessToken)
	})

	t.Run("access token also valid, passed", func(t *testing.T) {
		refreshToken := &models.RefreshToken{
			RefreshToken: newRefreshToken,
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
