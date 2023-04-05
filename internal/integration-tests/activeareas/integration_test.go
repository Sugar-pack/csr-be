package activeareas

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/active_areas"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
)

func TestIntegration_GetActiveAreas(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

	t.Run("get all active areas ok", func(t *testing.T) {
		params := active_areas.NewGetAllActiveAreasParamsWithContext(ctx)

		got, err := client.ActiveAreas.GetAllActiveAreas(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)
		assert.NotEmpty(t, got)
	})

	t.Run("get all active areas failed: no authorization", func(t *testing.T) {
		params := active_areas.NewGetAllActiveAreasParamsWithContext(ctx)

		_, err := client.ActiveAreas.GetAllActiveAreas(params, utils.AuthInfoFunc(nil))
		require.Error(t, err)

		errExp := active_areas.NewGetAllActiveAreasDefault(http.StatusUnauthorized)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("get all active areas failed: token contains an invalid number of segments", func(t *testing.T) {
		params := active_areas.NewGetAllActiveAreasParamsWithContext(ctx)
		dummyToken := utils.TokenNotExist

		_, err := client.ActiveAreas.GetAllActiveAreas(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, err)

		errExp := active_areas.NewGetAllActiveAreasDefault(http.StatusUnauthorized)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})
}
