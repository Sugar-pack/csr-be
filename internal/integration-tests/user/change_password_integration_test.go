package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
)

func TestIntegration_ChangePassword(t *testing.T) {
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
	newPassword := "newPassword"

	token := loginUser.GetPayload().AccessToken
	t.Run("non-valid old password", func(t *testing.T) {
		params := users.NewChangePasswordParamsWithContext(ctx)
		params.PasswordPatch = &models.PatchPasswordRequest{
			NewPassword: newPassword,
			OldPassword: "wrongPassword",
		}

		auth := utils.AuthInfoFunc(token)

		_, err = client.Users.ChangePassword(params, auth)
		require.Error(t, err)
	})

	t.Run("password changed", func(t *testing.T) {
		params := users.NewChangePasswordParamsWithContext(ctx)
		params.PasswordPatch = &models.PatchPasswordRequest{
			NewPassword: newPassword,
			OldPassword: p,
		}

		auth := utils.AuthInfoFunc(token)

		_, err = client.Users.ChangePassword(params, auth)
		require.NoError(t, err)

		_, err = utils.LoginUser(ctx, client, l, newPassword)
		require.NoError(t, err)
	})
}
