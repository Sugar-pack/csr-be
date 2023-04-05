package roles

import (
	"context"
	"net/http"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/roles"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_GetRoles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	loginUser := utils.AdminUserLogin(t)

	t.Run("get roles ok", func(t *testing.T) {
		params := roles.NewGetRolesParamsWithContext(ctx)
		token := loginUser.GetPayload().AccessToken

		res, err := client.Roles.GetRoles(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("get roles failed: no authorization", func(t *testing.T) {
		params := roles.NewGetRolesParamsWithContext(ctx)

		_, err := client.Roles.GetRoles(params, utils.AuthInfoFunc(nil))
		require.Error(t, err)

		errExp := roles.NewGetRolesDefault(http.StatusUnauthorized)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("get roles failed: token contains an invalid number of segments", func(t *testing.T) {
		params := roles.NewGetRolesParamsWithContext(ctx)
		dummyToken := utils.TokenNotExist

		_, err := client.Roles.GetRoles(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, err)

		errExp := roles.NewGetRolesDefault(http.StatusUnauthorized)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})
}
