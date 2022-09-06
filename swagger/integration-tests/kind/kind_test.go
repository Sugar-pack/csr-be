package kind

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/kinds"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/integration-tests/common"
)

var (
	testKindName        = gofakeit.Name()
	migrationKindNumber = 9
	testLogin           string
	testPassword        string
	auth                runtime.ClientAuthInfoWriterFunc
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Short() {
		ctx := context.Background()
		beClient := utils.SetupClient()

		var err error
		testLogin, testPassword, err = utils.GenerateLoginAndPassword()
		if err != nil {
			log.Fatalf("GenerateLoginAndPassword: %v", err)
		}
		_, err = utils.CreateUser(ctx, beClient, testLogin, testPassword)
		if err != nil {
			log.Fatalf("CreateUser: %v", err)
		}
		loginUser, err := utils.LoginUser(ctx, beClient, testLogin, testPassword)
		if err != nil {
			log.Fatalf("LoginUser: %v", err)
		}

		auth = utils.AuthInfoFunc(loginUser.GetPayload().AccessToken)

		os.Exit(m.Run())
	}
}

func TestIntegration_GetAllKind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	t.Run("Get All Kinds ok", func(t *testing.T) {
		params := kinds.NewGetAllKindsParamsWithContext(ctx)
		res, err := client.Kinds.GetAllKinds(params, auth)
		require.NoError(t, err)
		// migration has 9 kinds, expect this count
		want := migrationKindNumber
		got := len(res.Payload.Items)
		assert.Equal(t, want, got)
	})

	t.Run("Get All Kinds failed: wrong column", func(t *testing.T) {
		params := kinds.NewGetAllKindsParamsWithContext(ctx)
		name := kind.FieldMaxReservationUnits
		params.OrderColumn = &name
		_, gotErr := client.Kinds.GetAllKinds(params, auth)
		require.Error(t, gotErr)

		wantErr := kinds.NewGetAllKindsDefault(http.StatusUnprocessableEntity)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_CreateKind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	t.Run("Create New Kind failed: access failed", func(t *testing.T) {
		name := testKindName + "-test"
		maxTime := int64(24)
		maxUnits := int64(2)
		modelKind := &models.CreateNewKind{
			MaxReservationTime:  &maxTime,
			MaxReservationUnits: &maxUnits,
			Name:                &name,
		}
		token := utils.TokenNotExist
		_, gotErr := client.Kinds.CreateNewKind(kinds.NewCreateNewKindParamsWithContext(ctx).WithNewKind(modelKind), utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := kinds.NewCreateNewKindDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create New Kind failed: validation failed, kind model not provided, expect 422", func(t *testing.T) {
		_, err := client.Kinds.CreateNewKind(kinds.NewCreateNewKindParamsWithContext(ctx), auth)
		require.Error(t, err)

		var gotErr *kinds.CreateNewKindDefault
		require.True(t, errors.As(err, &gotErr))
		assert.Equal(t, http.StatusUnprocessableEntity, gotErr.Code())
	})

	t.Run("Create New Kind OK", func(t *testing.T) {
		// check won't pass for negative values.
		// values for date and units cannot be negative, but the manager sets up parameters, so we rely on manager
		name := testKindName + "-test"
		maxTime := int64(24)
		maxUnits := int64(2)
		modelKind := &models.CreateNewKind{
			MaxReservationTime:  &maxTime,
			MaxReservationUnits: &maxUnits,
			Name:                &name,
		}

		newKind, err := client.Kinds.CreateNewKind(kinds.NewCreateNewKindParamsWithContext(ctx).WithNewKind(modelKind), auth)
		require.NoError(t, err)

		require.NotNil(t, newKind.Payload.Data)
		assert.Equal(t, name, *newKind.Payload.Data.Name)
		assert.Equal(t, maxTime, newKind.Payload.Data.MaxReservationTime)
		assert.Equal(t, maxUnits, newKind.Payload.Data.MaxReservationUnits)
	})
}

func TestIntegration_GetKindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	t.Run("Get Kind By ID OK", func(t *testing.T) {
		name := testKindName + "-test2"
		maxTime := int64(24)
		maxUnits := int64(2)
		modelKind := &models.CreateNewKind{
			MaxReservationTime:  &maxTime,
			MaxReservationUnits: &maxUnits,
			Name:                &name,
		}

		newKind, err := client.Kinds.CreateNewKind(kinds.NewCreateNewKindParamsWithContext(ctx).WithNewKind(modelKind), auth)
		require.NoError(t, err)

		res, err := client.Kinds.GetKindByID(kinds.NewGetKindByIDParamsWithContext(ctx).WithKindID(newKind.Payload.Data.ID), auth)
		require.NoError(t, err)

		assert.Equal(t, newKind.Payload.Data, res.Payload.Data)
	})

	t.Run("Get Kind By ID failed: incorrect ID", func(t *testing.T) {
		_, gotErr := client.Kinds.GetKindByID(kinds.NewGetKindByIDParamsWithContext(ctx).WithKindID(-33), auth)
		require.Error(t, gotErr)

		wantErr := kinds.NewGetKindByIDDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{
			Message: "failed to get kind",
		}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Get Kind By ID failed: access failed", func(t *testing.T) {
		token := utils.TokenNotExist
		_, gotErr := client.Kinds.GetKindByID(kinds.NewGetKindByIDParamsWithContext(ctx).WithKindID(1), utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := kinds.NewGetKindByIDDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_EditKind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	name := testKindName + "-test3"
	maxTime := int64(24)
	maxUnits := int64(2)
	modelKind := &models.CreateNewKind{
		MaxReservationTime:  &maxTime,
		MaxReservationUnits: &maxUnits,
		Name:                &name,
	}

	newKind, err := client.Kinds.CreateNewKind(kinds.NewCreateNewKindParamsWithContext(ctx).WithNewKind(modelKind), auth)
	require.NoError(t, err)

	t.Run("Edit Kind By ID OK: patchKind parameters are empty, not changed", func(t *testing.T) {
		patchKind := &models.PatchKind{
			Name:                "",
			MaxReservationTime:  int64(0),
			MaxReservationUnits: int64(0),
		}
		res, err := client.Kinds.PatchKind(kinds.NewPatchKindParamsWithContext(ctx).WithKindID(newKind.Payload.Data.ID).WithPatchKind(patchKind), auth)
		require.NoError(t, err)

		assert.Equal(t, newKind.Payload.Data.Name, res.Payload.Data.Name)
		assert.Equal(t, newKind.Payload.Data.MaxReservationTime, res.Payload.Data.MaxReservationTime)
		assert.Equal(t, newKind.Payload.Data.MaxReservationUnits, res.Payload.Data.MaxReservationUnits)
	})

	t.Run("Edit Kind By ID Ok: patchKind parameters not empty, changed", func(t *testing.T) {
		patchKind := &models.PatchKind{
			MaxReservationTime:  int64(23),
			MaxReservationUnits: int64(3),
			Name:                "test kind name",
		}
		res, err := client.Kinds.PatchKind(kinds.NewPatchKindParamsWithContext(ctx).WithKindID(newKind.Payload.Data.ID).WithPatchKind(patchKind), auth)
		require.NoError(t, err)

		require.Equal(t, patchKind.Name, *res.Payload.Data.Name)
		require.Equal(t, patchKind.MaxReservationTime, res.Payload.Data.MaxReservationTime)
		require.Equal(t, patchKind.MaxReservationUnits, res.Payload.Data.MaxReservationUnits)
	})

	t.Run("Edit Kind By ID failed: ID incorrect", func(t *testing.T) {
		patchKind := &models.PatchKind{
			MaxReservationTime:  int64(23),
			MaxReservationUnits: int64(3),
			Name:                "test kind name",
		}
		_, gotErr := client.Kinds.PatchKind(kinds.NewPatchKindParamsWithContext(ctx).WithKindID(-33).WithPatchKind(patchKind), auth)
		require.Error(t, gotErr)

		wantErr := kinds.NewPatchKindDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{
			Message: "cant update kind",
		}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Edit Kind By ID failed: access failed", func(t *testing.T) {
		patchKind := &models.PatchKind{
			MaxReservationTime:  int64(23),
			MaxReservationUnits: int64(3),
			Name:                "test kind name",
		}
		token := utils.TokenNotExist

		_, gotErr := client.Kinds.PatchKind(kinds.NewPatchKindParamsWithContext(ctx).WithKindID(newKind.Payload.Data.ID).WithPatchKind(patchKind), utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := kinds.NewPatchKindDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Edit Kind By ID failed: validation failed, expect 422", func(t *testing.T) {
		_, gotErr := client.Kinds.PatchKind(kinds.NewPatchKindParamsWithContext(ctx).WithKindID(newKind.Payload.Data.ID), auth)
		require.Error(t, gotErr)

		wantErr := kinds.NewPatchKindDefault(http.StatusUnprocessableEntity)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_DeleteKind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	t.Run("Delete Kind By ID failed: not a validation error, kindID is not required in spec, expect 500", func(t *testing.T) {
		_, gotErr := client.Kinds.DeleteKind(kinds.NewDeleteKindParamsWithContext(ctx), auth)
		require.Error(t, gotErr)

		wantErr := kinds.NewDeleteKindDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "delete kind failed"}}
		assert.Equal(t, wantErr, gotErr)

		_, gotErr = client.Kinds.DeleteKind(kinds.NewDeleteKindParamsWithContext(ctx).WithKindID(-33), auth)
		require.Error(t, gotErr)

		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Kind By ID failed: access failed", func(t *testing.T) {
		token := utils.TokenNotExist
		_, gotErr := client.Kinds.DeleteKind(kinds.NewDeleteKindParamsWithContext(ctx).WithKindID(1), utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := kinds.NewDeleteKindDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Kind By ID OK", func(t *testing.T) {
		id := int64(migrationKindNumber + 1)
		res, err := client.Kinds.DeleteKind(kinds.NewDeleteKindParamsWithContext(ctx).WithKindID(id), auth)
		require.NoError(t, err)

		assert.Equal(t, "kind deleted", res.Payload)
	})
}
