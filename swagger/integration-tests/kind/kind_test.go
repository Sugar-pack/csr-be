package kind

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/kinds"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/integration-tests/common"
)

func TestIntegration_Kind(t *testing.T) {
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

	token := loginUser.GetPayload().AccessToken
	t.Run("register a new kind ok", func(t *testing.T) {
		params := kinds.NewGetAllKindsParamsWithContext(ctx)
		res, err := client.Kinds.GetAllKinds(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		j := res.GetPayload().Items[0].ID
		_ = j
		require.NoError(t, err)
	})
}
