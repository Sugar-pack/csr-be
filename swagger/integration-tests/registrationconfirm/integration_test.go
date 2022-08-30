package registrationconfirm

import (
	"context"
	"net/http"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/registration_confirm"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/integration-tests/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_SendConfirmLink(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = utils.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	t.Run("registration confirmation login success", func(t *testing.T) {
		// set env variable EMAIL_SENDER_IS_SEND_REQUIRED as false to not send the link in tests
		want := "Confirmation link was not sent to email, sending parameter was set to false and not required"
		params := registration_confirm.NewSendRegistrationConfirmLinkByLoginParamsWithContext(ctx)
		params.Login = &models.SendRegistrationConfirmLinkRequest{
			Data: &models.Login{
				Login: &l,
			},
		}

		res, err := client.RegistrationConfirm.SendRegistrationConfirmLinkByLogin(params)
		got := res.GetPayload()
		require.NoError(t, err)
		assert.Equal(t, want, string(got))
	})

	t.Run("registration confirmation failed: incorrect login", func(t *testing.T) {
		testLogin := utils.LoginNotExist
		params := registration_confirm.NewSendRegistrationConfirmLinkByLoginParamsWithContext(ctx)
		params.Login = &models.SendRegistrationConfirmLinkRequest{Data: &models.Login{
			Login: &testLogin,
		}}
		_, err = client.RegistrationConfirm.SendRegistrationConfirmLinkByLogin(params)
		require.Error(t, err)

		errExp := registration_confirm.NewSendRegistrationConfirmLinkByLoginDefault(http.StatusInternalServerError)
		errExp.Payload = &models.Error{
			Data: &models.ErrorData{Message: "Can't find this user, registration confirmation link wasn't send"},
		}
		assert.Equal(t, errExp, err)
	})
}
