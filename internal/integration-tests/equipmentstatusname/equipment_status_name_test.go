package equipmentstatusname

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	eqStatusName "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/equipment_status_name"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
)

func TestIntegration_GetStatuses(t *testing.T) {
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

	t.Run("Get List Statuses", func(t *testing.T) {
		params := eqStatusName.NewListEquipmentStatusNamesParamsWithContext(ctx)
		got, err := client.EquipmentStatusName.ListEquipmentStatusNames(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)
		// expect len 5 according to migration
		want := 5
		assert.Equal(t, want, len(got.GetPayload()))
	})

	t.Run("Get Status failed: invalid auth", func(t *testing.T) {
		params := eqStatusName.NewListEquipmentStatusNamesParamsWithContext(ctx)
		dummyToken := utils.TokenNotExist
		_, gotErr := client.EquipmentStatusName.ListEquipmentStatusNames(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewListEquipmentStatusNamesDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_GetStatus(t *testing.T) {
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

	t.Run("Get Status", func(t *testing.T) {
		params := eqStatusName.NewGetEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = 1
		res, err := client.EquipmentStatusName.GetEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		got := *res.Payload.Data.Name
		want := "in review"
		assert.Equal(t, want, got)
	})

	t.Run("Get Status failed: status unknown", func(t *testing.T) {
		params := eqStatusName.NewGetEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = -10
		_, gotErr := client.EquipmentStatusName.GetEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewGetEquipmentStatusNameDefault(500)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{
			Message: "can't get status",
		}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Get Status failed: invalid auth", func(t *testing.T) {
		// name = "review"
		params := eqStatusName.NewGetEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = 1
		dummyToken := utils.TokenNotExist
		_, gotErr := client.EquipmentStatusName.GetEquipmentStatusName(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewGetEquipmentStatusNameDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_PostStatus(t *testing.T) {
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

	t.Run("Post Status successfully", func(t *testing.T) {
		want := "test status"
		params := eqStatusName.NewPostEquipmentStatusNameParamsWithContext(ctx)
		params.Name = &models.EquipmentStatusName{
			Name: &want,
		}
		res, err := client.EquipmentStatusName.PostEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		got := *res.GetPayload().Data.Name
		assert.Equal(t, want, got)
	})

	t.Run("Post Status failed: post same status again", func(t *testing.T) {
		wantSameStatus := "test status"
		params := eqStatusName.NewPostEquipmentStatusNameParamsWithContext(ctx)
		params.Name = &models.EquipmentStatusName{
			Name: &wantSameStatus,
		}
		_, gotErr := client.EquipmentStatusName.PostEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewPostEquipmentStatusNameDefault(500)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "can't create status"}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Post Status failed: invalid auth", func(t *testing.T) {
		want := "new status"
		params := eqStatusName.NewPostEquipmentStatusNameParamsWithContext(ctx)
		params.Name = &models.EquipmentStatusName{
			Name: &want,
		}
		dummyToken := utils.TokenNotExist
		_, gotErr := client.EquipmentStatusName.PostEquipmentStatusName(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewPostEquipmentStatusNameDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_DeleteStatus(t *testing.T) {
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

	t.Run("Delete Status successfully", func(t *testing.T) {
		params := eqStatusName.NewDeleteEquipmentStatusNameParamsWithContext(ctx)
		// StatusID = 6 is "test status"
		params.StatusID = 6
		res, err := client.EquipmentStatusName.DeleteEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		want := "test status"
		got := *res.Payload.Data.Name
		assert.Equal(t, want, got)
	})

	t.Run("Delete Status failed: trying to delete same status", func(t *testing.T) {
		params := eqStatusName.NewDeleteEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = 6
		_, gotErr := client.EquipmentStatusName.DeleteEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewDeleteEquipmentStatusNameDefault(500)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "can't delete status"}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Status failed: trying to delete unknown status", func(t *testing.T) {
		params := eqStatusName.NewDeleteEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = -10
		_, gotErr := client.EquipmentStatusName.DeleteEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewDeleteEquipmentStatusNameDefault(500)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "can't delete status"}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Status failed: invalid auth", func(t *testing.T) {
		// name = "review"
		params := eqStatusName.NewDeleteEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = 1
		dummyToken := utils.TokenNotExist
		_, gotErr := client.EquipmentStatusName.DeleteEquipmentStatusName(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewDeleteEquipmentStatusNameDefault(500)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}
