package petkind

import (
	"context"
	"net/http"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/pet_kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_PetKind(t *testing.T) {
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

	t.Run("register a new pet kind ok", func(t *testing.T) {
		token := loginUser.GetPayload().AccessToken

		params := pet_kind.NewCreateNewPetKindParamsWithContext(ctx)
		params.NewPetKind = &models.PetKind{
			// should not provide ID
			Name: &petKindName,
		}

		want := pet_kind.NewCreateNewPetKindCreated()
		want.Payload = &models.PetKindResponse{
			Name: &petKindName,
		}

		got, err := client.PetKind.CreateNewPetKind(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)
		assert.NotEmpty(t, want.GetPayload().Name, got.GetPayload().Name)
	})

	t.Run("register a new pet kind failed: validation error", func(t *testing.T) {
		token := loginUser.GetPayload().AccessToken

		params := pet_kind.NewCreateNewPetKindParamsWithContext(ctx)
		params.NewPetKind = &models.PetKind{
			Name: nil,
		}

		_, err = client.PetKind.CreateNewPetKind(params, utils.AuthInfoFunc(token))
		require.Error(t, err)

		errExp := pet_kind.NewCreateNewPetKindDefault(http.StatusUnprocessableEntity)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("register a new pet kind failed: no authorization", func(t *testing.T) {
		params := pet_kind.NewCreateNewPetKindParamsWithContext(ctx)

		_, err = client.PetKind.CreateNewPetKind(params, utils.AuthInfoFunc(nil))
		require.Error(t, err)

		errExp := pet_kind.NewCreateNewPetKindDefault(http.StatusUnauthorized)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("register a new pet kind failed: token contains an invalid number of segments", func(t *testing.T) {
		params := pet_kind.NewCreateNewPetKindParamsWithContext(ctx)
		dummyToken := utils.TokenNotExist

		_, err = client.PetKind.CreateNewPetKind(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, err)

		errExp := pet_kind.NewCreateNewPetKindDefault(http.StatusInternalServerError)
		errExp.Payload = &models.Error{
			Data: nil,
		}
		assert.Equal(t, errExp, err)
	})
}

func getKindIDByName(ctx context.Context, client *client.Be, token *string, petKindName string) (*int64, error) {
	paramsGetAll := pet_kind.NewGetAllPetKindsParamsWithContext(ctx)
	petKinds, err := client.PetKind.GetAllPetKinds(paramsGetAll, utils.AuthInfoFunc(token))
	if err != nil {
		return nil, err
	}
	var petKindID *int64

	for _, petKind := range petKinds.GetPayload() {
		if *petKind.Name == petKindName {
			petKindID = petKind.ID
		}
	}
	return petKindID, nil
}
