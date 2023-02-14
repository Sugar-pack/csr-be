package petsize

import (
	"context"
	"net/http"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/pet_size"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	size = "test size"
	name = "test pet name"
)

func TestIntegration_PetSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

	t.Run("register a new pet size ok", func(t *testing.T) {

		params := pet_size.NewCreateNewPetSizeParamsWithContext(ctx)
		params.NewPetSize = &models.PetSize{
			// should not provide ID
			Name: &name,
			Size: &size,
		}

		want := pet_size.NewCreateNewPetSizeCreated()
		want.Payload = &models.PetSizeResponse{
			Name: &name,
			Size: &size,
		}

		got, err := client.PetSize.CreateNewPetSize(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)
		assert.NotEmpty(t, want.GetPayload().Name, got.GetPayload().Name)
	})

	t.Run("register a new pet kind failed: validation error", func(t *testing.T) {

		params := pet_size.NewCreateNewPetSizeParamsWithContext(ctx)
		params.NewPetSize = &models.PetSize{
			Name: nil,
		}

		_, err := client.PetSize.CreateNewPetSize(params, utils.AuthInfoFunc(token))
		require.Error(t, err)

		errExp := pet_size.NewCreateNewPetSizeDefault(http.StatusUnprocessableEntity)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("register a new pet kind failed: no authorization", func(t *testing.T) {
		params := pet_size.NewCreateNewPetSizeParamsWithContext(ctx)

		_, err := client.PetSize.CreateNewPetSize(params, utils.AuthInfoFunc(nil))
		require.Error(t, err)

		errExp := pet_size.NewCreateNewPetSizeDefault(http.StatusUnauthorized)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("register a new pet kind failed: token contains an invalid number of segments", func(t *testing.T) {
		params := pet_size.NewCreateNewPetSizeParamsWithContext(ctx)
		dummyToken := utils.TokenNotExist

		_, err := client.PetSize.CreateNewPetSize(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, err)

		errExp := pet_size.NewCreateNewPetSizeDefault(http.StatusUnauthorized)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})
}
