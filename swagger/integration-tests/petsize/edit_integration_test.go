package petsize

import (
	"context"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/pet_size"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/integration-tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_EditPetKind(t *testing.T) {
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

	petSizeID, err := getSizeIDByName(ctx, client, token, name)
	require.NoError(t, err)

	params := pet_size.NewEditPetSizeParamsWithContext(ctx)
	newSize := "new test size"
	newName := "new name"
	params.PetSizeID = *petSizeID
	params.EditPetSize = &models.PetSize{
		// should not provide ID
		Name: &newName,
		Size: &newSize,
	}

	kind, err := client.PetSize.EditPetSize(params, utils.AuthInfoFunc(token))
	require.NoError(t, err)

	assert.Equal(t, newName, *kind.GetPayload().Name)
	assert.Equal(t, newSize, *kind.GetPayload().Size)

	// revert changes back for delete function
	params.EditPetSize = &models.PetSize{
		Name: &name,
		Size: &size,
	}

	_, err = client.PetSize.EditPetSize(params, utils.AuthInfoFunc(token))
	assert.NoError(t, err)
}
