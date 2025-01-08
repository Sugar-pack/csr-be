package petsize

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/pet_size"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
)

func TestIntegration_GetAllPetSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

	params := pet_size.NewGetAllPetSizeParamsWithContext(ctx)

	_, err := client.PetSize.GetAllPetSize(params, utils.AuthInfoFunc(token))
	require.NoError(t, err)
}

func TestIntegration_GetPetSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

	t.Run("get pet size ok", func(t *testing.T) {

		petSizeID, err := getSizeIDByName(ctx, client, token, name)
		require.NoError(t, err)

		paramsGet := pet_size.NewGetPetSizeParamsWithContext(ctx)
		paramsGet.SetPetSizeID(*petSizeID)

		got, err := client.PetSize.GetPetSize(paramsGet, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		want := pet_size.NewGetPetSizeOK()
		want.Payload = &models.PetSizeResponse{
			ID:   petSizeID,
			Name: &name,
			Size: &size,
		}

		assert.Equal(t, got, want)
	})
}

func TestIntegration_DeletePetSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

	petSizeID, err := getSizeIDByName(ctx, client, token, name)
	require.NoError(t, err)

	params := pet_size.NewDeletePetSizeParamsWithContext(ctx)
	params.PetSizeID = *petSizeID

	res, err := client.PetSize.DeletePetSize(params, utils.AuthInfoFunc(token))
	require.NoError(t, err)

	assert.Equal(t, "pet size deleted", res.GetPayload())
}

func getSizeIDByName(ctx context.Context, client *client.Be, token *string, petSizeName string) (*int64, error) {
	paramsGetAll := pet_size.NewGetAllPetSizeParamsWithContext(ctx)
	allPetSize, err := client.PetSize.GetAllPetSize(paramsGetAll, utils.AuthInfoFunc(token))
	if err != nil {
		return nil, err
	}
	var petSizeID *int64

	for _, petSize := range allPetSize.GetPayload() {
		if *petSize.Name == petSizeName {
			petSizeID = petSize.ID
		}
	}
	return petSizeID, nil
}
