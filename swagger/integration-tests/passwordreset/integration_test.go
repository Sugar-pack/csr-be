package passwordreset

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/password_reset"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/integration-tests/common"
)

func TestIntegration_PasswordReset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	user, err := utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	t.Run("create password reset link by login ok", func(t *testing.T) {
		params := password_reset.NewSendLinkByLoginParamsWithContext(ctx)
		params.Login = &models.SendPasswordResetLinkRequest{Data: &models.Login{
			Login: user.Login,
		}}
		got, err := client.PasswordReset.SendLinkByLogin(params)
		require.NoError(t, err)

		want := &password_reset.SendLinkByLoginOK{
			Payload: models.PasswordResetResponse("Reset link sent"),
		}
		assert.Equal(t, want, got)
	})

	t.Run("create password reset link failed: login not provided", func(t *testing.T) {
		params := password_reset.NewSendLinkByLoginParamsWithContext(ctx)
		login := ""
		params.Login = &models.SendPasswordResetLinkRequest{Data: &models.Login{
			Login: &login,
		}}
		_, err = client.PasswordReset.SendLinkByLogin(params)
		require.Error(t, err)

		errExp := password_reset.NewSendLinkByLoginDefault(http.StatusBadRequest)
		errExp.Payload = &models.Error{
			Data: &models.ErrorData{Message: "Login is required"},
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("create password reset link failed: login not exist, 500", func(t *testing.T) {
		params := password_reset.NewSendLinkByLoginParamsWithContext(ctx)
		login := utils.LoginNotExist
		params.Login = &models.SendPasswordResetLinkRequest{Data: &models.Login{
			Login: &login,
		}}
		_, err = client.PasswordReset.SendLinkByLogin(params)
		require.Error(t, err)

		errExp := password_reset.NewSendLinkByLoginDefault(http.StatusInternalServerError)
		errExp.Payload = &models.Error{
			Data: &models.ErrorData{
				Message: "Can't send reset password link. Please try again later",
			},
		}
		assert.Equal(t, errExp, err)
	})
}

func TestIntegration_PasswordResetGetLink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	t.Run("get password reset link by login failed: error while getting token", func(t *testing.T) {
		params := password_reset.NewGetPasswordResetLinkParamsWithContext(ctx)
		params.Token = utils.TokenNotExist
		_, err = client.PasswordReset.GetPasswordResetLink(params)
		assert.Error(t, err)

		errExp := password_reset.NewGetPasswordResetLinkDefault(http.StatusInternalServerError)
		errExp.Payload = &models.Error{Data: &models.ErrorData{
			Message: "Failed to verify token. Please try again later",
		}}
		assert.Equal(t, errExp, err)
	})
}
