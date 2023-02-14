package petsize

import (
	"context"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/pet_size"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_EditPetSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

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

	petSize, err := client.PetSize.EditPetSize(params, utils.AuthInfoFunc(token))
	require.NoError(t, err)

	assert.Equal(t, newName, *petSize.GetPayload().Name)
	assert.Equal(t, newSize, *petSize.GetPayload().Size)

	// revert changes back for delete function
	params.EditPetSize = &models.PetSize{
		Name: &name,
		Size: &size,
	}

	_, err = client.PetSize.EditPetSize(params, utils.AuthInfoFunc(token))
	assert.NoError(t, err)
}
