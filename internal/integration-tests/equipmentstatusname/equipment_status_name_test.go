package equipmentstatusname

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	eqStatusName "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/equipment_status_name"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func TestIntegration_GetStatuses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

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

		wantErr := eqStatusName.NewListEquipmentStatusNamesDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_GetStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

	t.Run("Get Status", func(t *testing.T) {
		params := eqStatusName.NewGetEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = 1
		res, err := client.EquipmentStatusName.GetEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		got := *res.Payload.Data.Name
		want := domain.EquipmentStatusAvailable
		assert.Equal(t, want, got)
	})

	t.Run("Get Status failed: status unknown", func(t *testing.T) {
		params := eqStatusName.NewGetEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = -10
		_, gotErr := client.EquipmentStatusName.GetEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewGetEquipmentStatusNameDefault(500)
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrGetEqStatus,
			Details: "ent: equipment_status_name not found",
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Get Status failed: invalid auth", func(t *testing.T) {
		// name = "review"
		params := eqStatusName.NewGetEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = 1
		dummyToken := utils.TokenNotExist
		_, gotErr := client.EquipmentStatusName.GetEquipmentStatusName(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewGetEquipmentStatusNameDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_PostStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

	t.Run("Post Status successfully", func(t *testing.T) {
		want := "test status"
		params := eqStatusName.NewPostEquipmentStatusNameParamsWithContext(ctx)
		params.Name = &models.NewEquipmentStatusName{
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
		params.Name = &models.NewEquipmentStatusName{
			Name: &wantSameStatus,
		}
		_, gotErr := client.EquipmentStatusName.PostEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewPostEquipmentStatusNameDefault(http.StatusInternalServerError)
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrCreateEqStatus,
			Details: "ent: constraint failed: ERROR: duplicate key value violates unique constraint \"equipment_status_names_name_key\" (SQLSTATE 23505)",
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Post Status failed: invalid auth", func(t *testing.T) {
		want := "new status"
		params := eqStatusName.NewPostEquipmentStatusNameParamsWithContext(ctx)
		params.Name = &models.NewEquipmentStatusName{
			Name: &want,
		}
		dummyToken := utils.TokenNotExist
		_, gotErr := client.EquipmentStatusName.PostEquipmentStatusName(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewPostEquipmentStatusNameDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_DeleteStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	auth := utils.AdminUserLogin(t)
	token := auth.GetPayload().AccessToken

	t.Run("Delete Status successfully", func(t *testing.T) {
		params := eqStatusName.NewDeleteEquipmentStatusNameParamsWithContext(ctx)
		// StatusID = 6 is "test status"
		params.StatusID = 6 // todo: get statusID from database
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

		wantErr := eqStatusName.NewDeleteEquipmentStatusNameDefault(http.StatusInternalServerError)
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrDeleteEqStatus,
			Details: "ent: equipment_status_name not found",
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Status failed: trying to delete unknown status", func(t *testing.T) {
		params := eqStatusName.NewDeleteEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = -10
		_, gotErr := client.EquipmentStatusName.DeleteEquipmentStatusName(params, utils.AuthInfoFunc(token))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewDeleteEquipmentStatusNameDefault(http.StatusInternalServerError)
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrDeleteEqStatus,
			Details: "ent: equipment_status_name not found",
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Status failed: invalid auth", func(t *testing.T) {
		// name = "review"
		params := eqStatusName.NewDeleteEquipmentStatusNameParamsWithContext(ctx)
		params.StatusID = 1
		dummyToken := utils.TokenNotExist
		_, gotErr := client.EquipmentStatusName.DeleteEquipmentStatusName(params, utils.AuthInfoFunc(&dummyToken))
		require.Error(t, gotErr)

		wantErr := eqStatusName.NewDeleteEquipmentStatusNameDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}
