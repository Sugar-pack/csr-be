package petkind

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/pet_kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
)

var (
	petKindName = "попугай"
)

func TestIntegration_GetAllPetKind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	loginUser := utils.AdminUserLogin(t)
	token := loginUser.GetPayload().AccessToken
	params := pet_kind.NewGetAllPetKindsParamsWithContext(ctx)

	_, err := client.PetKind.GetAllPetKinds(params, utils.AuthInfoFunc(token))
	require.NoError(t, err)
}

func TestIntegration_GetPetKind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	loginUser := utils.AdminUserLogin(t)

	t.Run("get pet kind ok", func(t *testing.T) {
		token := loginUser.GetPayload().AccessToken

		petKindID, err := getKindIDByName(ctx, client, token, petKindName)
		require.NoError(t, err)

		paramsGet := pet_kind.NewGetPetKindParamsWithContext(ctx)
		paramsGet.SetPetKindID(*petKindID)

		got, err := client.PetKind.GetPetKind(paramsGet, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		want := pet_kind.NewGetPetKindOK()
		want.Payload = &models.PetKindResponse{
			ID:   petKindID,
			Name: &petKindName,
		}

		assert.Equal(t, got, want)
	})
}

func TestIntegration_EditPetKind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	loginUser := utils.AdminUserLogin(t)

	token := loginUser.GetPayload().AccessToken

	petKindID, err := getKindIDByName(ctx, client, token, petKindName)
	require.NoError(t, err)

	params := pet_kind.NewEditPetKindParamsWithContext(ctx)
	petKind := "динозавр"

	params.PetKindID = *petKindID
	params.EditPetKind = &models.PetKind{
		Name: &petKind,
	}

	kind, err := client.PetKind.EditPetKind(params, utils.AuthInfoFunc(token))
	require.NoError(t, err)

	assert.Equal(t, petKind, *kind.GetPayload().Name)

	// revert changes for delete function
	params.EditPetKind = &models.PetKind{
		Name: &petKindName,
	}

	_, err = client.PetKind.EditPetKind(params, utils.AuthInfoFunc(token))
	assert.NoError(t, err)
}

func TestIntegration_DeletePetKind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	loginUser := utils.AdminUserLogin(t)

	token := loginUser.GetPayload().AccessToken

	petKindID, err := getKindIDByName(ctx, client, token, petKindName)
	require.NoError(t, err)

	params := pet_kind.NewDeletePetKindParamsWithContext(ctx)
	params.PetKindID = *petKindID

	kind, err := client.PetKind.DeletePetKind(params, utils.AuthInfoFunc(token))
	require.NoError(t, err)

	assert.Equal(t, kind.GetPayload(), "pet kind deleted")
}
