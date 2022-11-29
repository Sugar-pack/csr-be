package category

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

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/category"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/categories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
)

var (
	testCategoryName        = gofakeit.Name()
	migrationCategoryNumber = 9
	testLogin               string
	testPassword            string
	auth                    runtime.ClientAuthInfoWriterFunc
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

func TestIntegration_GetAllCategories(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	t.Run("Get All Categories ok", func(t *testing.T) {
		params := categories.NewGetAllCategoriesParamsWithContext(ctx)
		res, err := client.Categories.GetAllCategories(params, auth)
		require.NoError(t, err)
		// migration has 9 categories, expect this count
		want := migrationCategoryNumber
		got := len(res.Payload.Items)
		assert.Equal(t, want, got)
	})

	t.Run("Get All Categories failed: wrong column", func(t *testing.T) {
		params := categories.NewGetAllCategoriesParamsWithContext(ctx)
		name := category.FieldMaxReservationUnits
		params.OrderColumn = &name
		_, gotErr := client.Categories.GetAllCategories(params, auth)
		require.Error(t, gotErr)

		wantErr := categories.NewGetAllCategoriesDefault(http.StatusUnprocessableEntity)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_CreateCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	t.Run("Create New Category failed: access failed", func(t *testing.T) {
		name := testCategoryName + "-test"
		maxTime := int64(24)
		maxUnits := int64(2)
		modelCategory := &models.CreateNewCategory{
			MaxReservationTime:  &maxTime,
			MaxReservationUnits: &maxUnits,
			Name:                &name,
		}
		token := utils.TokenNotExist
		_, gotErr := client.Categories.CreateNewCategory(categories.NewCreateNewCategoryParamsWithContext(ctx).WithNewCategory(modelCategory), utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := categories.NewCreateNewCategoryDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create New Category failed: validation failed, category model not provided, expect 422", func(t *testing.T) {
		_, err := client.Categories.CreateNewCategory(categories.NewCreateNewCategoryParamsWithContext(ctx), auth)
		require.Error(t, err)

		var gotErr *categories.CreateNewCategoryDefault
		require.True(t, errors.As(err, &gotErr))
		assert.Equal(t, http.StatusUnprocessableEntity, gotErr.Code())
	})

	t.Run("Create New Category OK", func(t *testing.T) {
		// check won't pass for negative values.
		// values for date and units cannot be negative, but the manager sets up parameters, so we rely on manager
		name := testCategoryName + "-test"
		maxTime := int64(24)
		maxUnits := int64(2)
		hasSubcats := true
		modelCategory := &models.CreateNewCategory{
			MaxReservationTime:  &maxTime,
			MaxReservationUnits: &maxUnits,
			Name:                &name,
			HasSubcategory:      &hasSubcats,
		}

		newCategory, err := client.Categories.CreateNewCategory(categories.NewCreateNewCategoryParamsWithContext(ctx).
			WithNewCategory(modelCategory), auth)
		require.NoError(t, err)

		require.NotNil(t, newCategory.Payload.Data)
		assert.Equal(t, name, *newCategory.Payload.Data.Name)
		assert.Equal(t, maxTime, *newCategory.Payload.Data.MaxReservationTime)
		assert.Equal(t, maxUnits, *newCategory.Payload.Data.MaxReservationUnits)
		assert.Equal(t, hasSubcats, *newCategory.Payload.Data.HasSubcategory)
	})
}

func TestIntegration_GetCategoryByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	t.Run("Get Category By ID OK", func(t *testing.T) {
		name := testCategoryName + "-test2"
		maxTime := int64(24)
		maxUnits := int64(2)
		hasSubcat := true
		modelCategory := &models.CreateNewCategory{
			MaxReservationTime:  &maxTime,
			MaxReservationUnits: &maxUnits,
			Name:                &name,
			HasSubcategory:      &hasSubcat,
		}

		newCategory, err := client.Categories.CreateNewCategory(categories.NewCreateNewCategoryParamsWithContext(ctx).
			WithNewCategory(modelCategory), auth)
		require.NoError(t, err)

		res, err := client.Categories.GetCategoryByID(categories.NewGetCategoryByIDParamsWithContext(ctx).
			WithCategoryID(*newCategory.Payload.Data.ID), auth)
		require.NoError(t, err)

		assert.Equal(t, newCategory.Payload.Data, res.Payload.Data)
	})

	t.Run("Get Category By ID failed: incorrect ID", func(t *testing.T) {
		_, gotErr := client.Categories.GetCategoryByID(categories.NewGetCategoryByIDParamsWithContext(ctx).WithCategoryID(-33), auth)
		require.Error(t, gotErr)

		wantErr := categories.NewGetCategoryByIDDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{
			Message: "failed to get category",
		}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Get Category By ID failed: access failed", func(t *testing.T) {
		token := utils.TokenNotExist
		_, gotErr := client.Categories.GetCategoryByID(categories.NewGetCategoryByIDParamsWithContext(ctx).WithCategoryID(1), utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := categories.NewGetCategoryByIDDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_EditCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	name := testCategoryName + "-test3"
	maxTime := int64(24)
	maxUnits := int64(2)
	hasSubcats := false
	modelCategory := &models.CreateNewCategory{
		MaxReservationTime:  &maxTime,
		MaxReservationUnits: &maxUnits,
		Name:                &name,
		HasSubcategory:      &hasSubcats,
	}

	newCategory, err := client.Categories.CreateNewCategory(categories.NewCreateNewCategoryParamsWithContext(ctx).
		WithNewCategory(modelCategory), auth)
	require.NoError(t, err)

	t.Run("Edit Category By ID OK: updateCategory parameters are same, not changed", func(t *testing.T) {
		updateCategory := &models.UpdateCategoryRequest{
			MaxReservationTime:  &maxTime,
			MaxReservationUnits: &maxUnits,
			Name:                &name,
			HasSubcategory:      &hasSubcats,
		}
		res, err := client.Categories.UpdateCategory(categories.NewUpdateCategoryParamsWithContext(ctx).
			WithCategoryID(*newCategory.Payload.Data.ID).WithUpdateCategory(updateCategory), auth)
		require.NoError(t, err)

		assert.Equal(t, newCategory.Payload.Data.Name, res.Payload.Data.Name)
		assert.Equal(t, newCategory.Payload.Data.MaxReservationTime, res.Payload.Data.MaxReservationTime)
		assert.Equal(t, newCategory.Payload.Data.MaxReservationUnits, res.Payload.Data.MaxReservationUnits)
		require.Equal(t, newCategory.Payload.Data.HasSubcategory, res.Payload.Data.HasSubcategory)
	})

	t.Run("Edit Category By ID Ok: updateCategory parameters not empty, changed", func(t *testing.T) {
		updName := "test"
		updMaxTime := int64(23)
		updMaxUnits := int64(3)
		updHasSubcats := true
		updateCategory := &models.UpdateCategoryRequest{
			Name:                &updName,
			MaxReservationTime:  &updMaxTime,
			MaxReservationUnits: &updMaxUnits,
			HasSubcategory:      &updHasSubcats,
		}
		res, err := client.Categories.UpdateCategory(categories.NewUpdateCategoryParamsWithContext(ctx).
			WithCategoryID(*newCategory.Payload.Data.ID).WithUpdateCategory(updateCategory), auth)
		require.NoError(t, err)

		require.Equal(t, updateCategory.Name, res.Payload.Data.Name)
		require.Equal(t, updateCategory.MaxReservationTime, res.Payload.Data.MaxReservationTime)
		require.Equal(t, updateCategory.MaxReservationUnits, res.Payload.Data.MaxReservationUnits)
		require.Equal(t, updateCategory.HasSubcategory, res.Payload.Data.HasSubcategory)
	})

	t.Run("Edit Category By ID failed: ID incorrect", func(t *testing.T) {
		updName := "test category name"
		updMaxTime := int64(23)
		updMaxUnits := int64(3)
		updHasSubcats := true
		updateCategory := &models.UpdateCategoryRequest{
			Name:                &updName,
			MaxReservationTime:  &updMaxTime,
			MaxReservationUnits: &updMaxUnits,
			HasSubcategory:      &updHasSubcats,
		}
		_, gotErr := client.Categories.UpdateCategory(categories.NewUpdateCategoryParamsWithContext(ctx).
			WithCategoryID(-33).WithUpdateCategory(updateCategory), auth)
		require.Error(t, gotErr)

		wantErr := categories.NewUpdateCategoryDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{
			Message: "cant update category",
		}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Edit Category By ID failed: access failed", func(t *testing.T) {
		updName := "test category name"
		updMaxTime := int64(23)
		updMaxUnits := int64(3)
		updateCategory := &models.UpdateCategoryRequest{
			Name:                &updName,
			MaxReservationTime:  &updMaxTime,
			MaxReservationUnits: &updMaxUnits,
		}
		token := utils.TokenNotExist

		_, gotErr := client.Categories.UpdateCategory(categories.NewUpdateCategoryParamsWithContext(ctx).
			WithCategoryID(*newCategory.Payload.Data.ID).WithUpdateCategory(updateCategory), utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := categories.NewUpdateCategoryDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Edit Category By ID failed: validation failed, expect 422", func(t *testing.T) {
		_, gotErr := client.Categories.UpdateCategory(categories.NewUpdateCategoryParamsWithContext(ctx).
			WithCategoryID(*newCategory.Payload.Data.ID), auth)
		require.Error(t, gotErr)

		wantErr := categories.NewUpdateCategoryDefault(http.StatusUnprocessableEntity)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_DeleteCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	t.Run("Delete Category By ID failed: not a validation error, categoryID is not required in spec, expect 500", func(t *testing.T) {
		_, gotErr := client.Categories.DeleteCategory(categories.NewDeleteCategoryParamsWithContext(ctx), auth)
		require.Error(t, gotErr)

		wantErr := categories.NewDeleteCategoryDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "delete category failed"}}
		assert.Equal(t, wantErr, gotErr)

		_, gotErr = client.Categories.DeleteCategory(categories.NewDeleteCategoryParamsWithContext(ctx).WithCategoryID(-33), auth)
		require.Error(t, gotErr)

		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Category By ID failed: access failed", func(t *testing.T) {
		token := utils.TokenNotExist
		_, gotErr := client.Categories.DeleteCategory(categories.NewDeleteCategoryParamsWithContext(ctx).WithCategoryID(1), utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := categories.NewDeleteCategoryDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Category By ID OK", func(t *testing.T) {
		id := int64(migrationCategoryNumber + 1)
		res, err := client.Categories.DeleteCategory(categories.NewDeleteCategoryParamsWithContext(ctx).
			WithCategoryID(id), auth)
		require.NoError(t, err)

		assert.Equal(t, "category deleted", res.Payload)
	})
}
