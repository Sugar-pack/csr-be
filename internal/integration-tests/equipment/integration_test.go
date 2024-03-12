package equipment

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/categories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/equipment"
	eqStatusName "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/equipment_status_name"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/pet_kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/pet_size"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/photos"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/subcategories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func TestIntegration_CreateEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	tokens := utils.AdminUserLogin(t)
	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

	t.Run("Create Equipment", func(t *testing.T) {
		params := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
		model, err := setParameters(ctx, client, auth)
		require.NoError(t, err)

		params.NewEquipment = model

		res, err := client.Equipment.CreateNewEquipment(params, auth)
		require.NoError(t, err)

		// location returned as nil, in discussion we decided that this parameter will have more that one values
		// for now it is not handled
		// todo: uncomment string below when it's handled properly
		assert.Equal(t, model.Category, res.Payload.Category)
		assert.Equal(t, model.CompensationCost, res.Payload.CompensationCost)
		assert.Equal(t, model.Condition, res.Payload.Condition)
		assert.Equal(t, model.Description, res.Payload.Description)
		assert.Equal(t, model.InventoryNumber, res.Payload.InventoryNumber)
		assert.Equal(t, model.Category, res.Payload.Category)
		//assert.Equal(t, location, *res.Payload.Location)
		assert.Equal(t, model.MaximumDays, res.Payload.MaximumDays)
		assert.Equal(t, model.Name, res.Payload.Name)
		assert.Equal(t, model.PetSize, res.Payload.PetSize)
		assert.Contains(t, *res.Payload.PhotoID, *model.PhotoID)
		assert.Equal(t, model.ReceiptDate, res.Payload.ReceiptDate)
		assert.Equal(t, model.Status, res.Payload.Status)
		assert.Equal(t, model.Supplier, res.Payload.Supplier)
		assert.Equal(t, model.TechnicalIssues, res.Payload.TechnicalIssues)
		assert.Equal(t, model.Title, res.Payload.Title)
	})

	t.Run("Create Equipment failed: 422 status code error, description and name fields have a number of characters greater than the limit ",
		func(t *testing.T) {
			params := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
			model, err := setParameters(ctx, client, auth)
			require.NoError(t, err)

			// name field tests:
			// max length of name field: 100 characters
			name, err := utils.GenerateRandomString(101)
			require.NoError(t, err)
			model.Name = &name
			params.NewEquipment = model

			_, err = client.Equipment.CreateNewEquipment(params, auth)
			require.Error(t, err)

			name, err = utils.GenerateRandomString(99)
			require.NoError(t, err)
			model.Name = &name
			params.NewEquipment = model

			_, err = client.Equipment.CreateNewEquipment(params, auth)
			require.NoError(t, err)

			// description field tests:
			// max length of description field: 255 characters
			model, err = setParameters(ctx, client, auth)
			require.NoError(t, err)
			description, err := utils.GenerateRandomString(256)
			require.NoError(t, err)
			model.Description = &description

			params.NewEquipment = model
			_, err = client.Equipment.CreateNewEquipment(params, auth)
			require.Error(t, err)

			description, err = utils.GenerateRandomString(254)
			require.NoError(t, err)
			model.Description = &description
			params.NewEquipment = model

			_, err = client.Equipment.CreateNewEquipment(params, auth)
			require.NoError(t, err)
		})

	t.Run("Create Equipment failed: foreign key constraint error", func(t *testing.T) {
		params := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
		model, err := setParameters(ctx, client, auth)
		require.NoError(t, err)

		id := ""
		model.PhotoID = &id
		params.NewEquipment = model

		_, gotErr := client.Equipment.CreateNewEquipment(params, auth)
		require.Error(t, gotErr)

		wantErr := equipment.NewCreateNewEquipmentDefault(500)
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrCreateEquipment,
			Details: "ent: constraint failed: ERROR: insert or update on table \"equipment\" violates foreign key constraint \"equipment_photo_equipments_fkey\" (SQLSTATE 23503)",
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create Equipment failed: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
		token := utils.TokenNotExist
		model, err := setParameters(ctx, client, auth)
		require.NoError(t, err)

		params.NewEquipment = model

		_, gotErr := client.Equipment.CreateNewEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewCreateNewEquipmentDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_GetAllEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	tokens := utils.AdminUserLogin(t)

	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

	t.Run("Get All Equipment", func(t *testing.T) {
		params := equipment.NewGetAllEquipmentParamsWithContext(ctx)

		res, err := client.Equipment.GetAllEquipment(params, auth)
		require.NoError(t, err)
		assert.NotZero(t, len(res.Payload.Items))
	})

	t.Run("Get All Equipment: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewGetAllEquipmentParamsWithContext(ctx)
		token := utils.TokenNotExist

		_, gotErr := client.Equipment.GetAllEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewGetAllEquipmentDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_GetEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	tokens := utils.AdminUserLogin(t)

	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)

	created, err := createEquipment(ctx, client, auth, model)
	require.NoError(t, err)

	t.Run("Get Equipment", func(t *testing.T) {
		params := equipment.NewGetEquipmentParamsWithContext(ctx)

		params.EquipmentID = *created.Payload.ID
		res, err := client.Equipment.GetEquipment(params, auth)
		require.NoError(t, err)

		// location returned as nil, in discussion we decided that this parameter will have more that one values
		// for now it is not handled
		// todo: uncomment string below when it's handled properly
		assert.Equal(t, model.Category, res.Payload.Category)
		assert.Equal(t, model.CompensationCost, res.Payload.CompensationCost)
		assert.Equal(t, model.Condition, res.Payload.Condition)
		assert.Equal(t, model.Description, res.Payload.Description)
		assert.Equal(t, model.InventoryNumber, res.Payload.InventoryNumber)
		assert.Equal(t, model.Category, res.Payload.Category)
		assert.Equal(t, model.MaximumDays, res.Payload.MaximumDays)
		assert.Equal(t, model.Name, res.Payload.Name)
		//assert.Equal(t, model.Location, res.Payload.Location)
		assert.Equal(t, model.PetSize, res.Payload.PetSize)
		assert.Contains(t, *res.Payload.PhotoID, *model.PhotoID)
		assert.Equal(t, model.ReceiptDate, res.Payload.ReceiptDate)
		assert.Equal(t, model.Status, res.Payload.Status)
		assert.Equal(t, model.Supplier, res.Payload.Supplier)
		assert.Equal(t, model.TechnicalIssues, res.Payload.TechnicalIssues)
		assert.Equal(t, model.Title, res.Payload.Title)
	})

	t.Run("Get Equipment failed: passed invalid equipment id", func(t *testing.T) {
		params := equipment.NewGetEquipmentParamsWithContext(ctx)
		require.NoError(t, err)

		params.EquipmentID = int64(-10)
		_, gotErr := client.Equipment.GetEquipment(params, auth)
		require.Error(t, gotErr)

		wantErr := equipment.NewGetEquipmentDefault(http.StatusInternalServerError)
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrGetEquipment,
			Details: "ent: equipment not found",
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Get Equipment failed: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewGetEquipmentParamsWithContext(ctx)
		require.NoError(t, err)

		params.EquipmentID = *created.Payload.ID
		token := utils.TokenNotExist
		_, gotErr := client.Equipment.GetEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewGetEquipmentDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_FindEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	tokens := utils.AdminUserLogin(t)

	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)

	_, err = createEquipment(ctx, client, auth, model)
	require.NoError(t, err)

	t.Run("Find Equipment", func(t *testing.T) {
		params := equipment.NewFindEquipmentParamsWithContext(ctx)
		params.FindEquipment = &models.EquipmentFilter{
			Category: *model.Category,
		}
		res, err := client.Equipment.FindEquipment(params, auth)
		require.NoError(t, err)

		assert.NotZero(t, *res.Payload.Total)
		for _, item := range res.Payload.Items {
			assert.Equal(t, model.Category, item.Category)
		}
	})

	t.Run("Find Equipment: limit = 1", func(t *testing.T) {
		params := equipment.NewFindEquipmentParamsWithContext(ctx)
		params.FindEquipment = &models.EquipmentFilter{
			Category: *model.Category,
		}
		limit := int64(1)
		params.WithLimit(&limit)

		res, err := client.Equipment.FindEquipment(params, auth)
		require.NoError(t, err)

		assert.Equal(t, int(limit), len(res.Payload.Items))
	})

	t.Run("Find Equipment failed: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewFindEquipmentParamsWithContext(ctx)
		params.FindEquipment = &models.EquipmentFilter{
			Category: *model.Category,
		}

		token := utils.TokenNotExist
		_, gotErr := client.Equipment.FindEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewFindEquipmentDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Find Equipment: unknown parameters, zero items found", func(t *testing.T) {
		params := equipment.NewFindEquipmentParamsWithContext(ctx)
		params.FindEquipment = &models.EquipmentFilter{
			TermsOfUse: "unknown category",
		}

		res, gotErr := client.Equipment.FindEquipment(params, auth)
		require.NoError(t, gotErr)

		assert.Zero(t, len(res.Payload.Items))
	})
}

func TestIntegration_EditEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	tokens := utils.AdminUserLogin(t)

	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)

	created, err := createEquipment(ctx, client, auth, model)
	require.NoError(t, err)

	t.Run("Edit Equipment description", func(t *testing.T) {
		desc := "new description"
		model.Description = &desc
		params := equipment.NewEditEquipmentParamsWithContext(ctx).WithEquipmentID(*created.Payload.ID).
			WithEditEquipment(model)

		res, err := client.Equipment.EditEquipment(params, auth)
		require.NoError(t, err)

		assert.Equal(t, desc, *res.Payload.Description)
	})

	t.Run("Edit Equipment description failed: wrong Equipment ID", func(t *testing.T) {
		desc := "new description"
		model.Description = &desc
		params := equipment.NewEditEquipmentParamsWithContext(ctx).WithEquipmentID(int64(-10)).
			WithEditEquipment(model)

		_, gotErr := client.Equipment.EditEquipment(params, auth)
		require.Error(t, gotErr)

		wantErr := equipment.NewEditEquipmentDefault(http.StatusInternalServerError)
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrUpdateEquipment,
			Details: "ent: equipment not found",
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Edit Equipment failed: authorization error 500 Invalid token", func(t *testing.T) {
		params := equipment.NewEditEquipmentParamsWithContext(ctx)
		require.NoError(t, err)

		params.EquipmentID = *created.Payload.ID
		token := utils.TokenNotExist
		_, gotErr := client.Equipment.EditEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewEditEquipmentDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_ArchiveEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	tokens := utils.AdminUserLogin(t)

	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)

	created, err := createEquipment(ctx, client, auth, model)
	require.NoError(t, err)

	t.Run("Archive Equipment", func(t *testing.T) {
		params := equipment.NewArchiveEquipmentParamsWithContext(ctx).WithEquipmentID(*created.Payload.ID)
		res, gotError := client.Equipment.ArchiveEquipment(params, auth)
		require.NoError(t, gotError)

		require.True(t, res.IsCode(http.StatusNoContent))
	})

	t.Run("Archive Equipment failed: equipment not found", func(t *testing.T) {
		params := equipment.NewArchiveEquipmentParamsWithContext(ctx).WithEquipmentID(-1)
		resp, gotErr := client.Equipment.ArchiveEquipment(params, auth)
		require.Error(t, gotErr)
		fmt.Print(resp)

		wantErr := equipment.NewArchiveEquipmentNotFound()
		codeExp := int32(http.StatusNotFound)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrEquipmentNotFound,
		}

		require.Equal(t, wantErr, gotErr)
	})

	t.Run("Archive Equipment with active orders", func(t *testing.T) {
		orStartDate, orEndDate := time.Now(), time.Now().AddDate(0, 0, 1)
		orderID, err := createOrder(ctx, client, auth, created.Payload.ID, orStartDate, orEndDate)
		require.NoError(t, err)
		params := equipment.NewArchiveEquipmentParamsWithContext(ctx).WithEquipmentID(*created.Payload.ID)
		var res *equipment.ArchiveEquipmentNoContent
		res, err = client.Equipment.ArchiveEquipment(params, auth)
		require.NoError(t, err)
		require.True(t, res.IsCode(http.StatusNoContent))
		ok, err := checkOrderStatus(ctx, client, auth, orderID, domain.OrderStatusClosed)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("Archive Equipment failed: auth failed", func(t *testing.T) {
		params := equipment.NewArchiveEquipmentParamsWithContext(ctx).WithEquipmentID(*created.Payload.ID)
		token := utils.TokenNotExist

		_, gotErr := client.Equipment.ArchiveEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewArchiveEquipmentDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})

	// todo: test for archive equipment with non-default status
}

func TestIntegration_BlockEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()
	startDate, endDate := strfmt.DateTime(time.Now().AddDate(0, 0, 1)), strfmt.DateTime(time.Now().AddDate(0, 0, 10))

	tokens := utils.AdminUserLogin(t)
	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)
	eq, err := createEquipment(ctx, client, auth, model)
	require.NoError(t, err)

	orStartDate, orEndDate := time.Now().AddDate(0, 0, 2), time.Now().AddDate(0, 0, 3)
	firstOrderID, err := createOrder(ctx, client, auth, eq.Payload.ID, orStartDate, orEndDate)
	require.NoError(t, err)
	require.NotNil(t, firstOrderID)

	orStartDate, orEndDate = time.Now().AddDate(0, 0, 4), time.Now().AddDate(0, 0, 5)
	secondOrderID, err := createOrder(ctx, client, auth, eq.Payload.ID, orStartDate, orEndDate)
	require.NoError(t, err)
	require.NotNil(t, secondOrderID)

	t.Run("Block Equipment with active orders", func(t *testing.T) {
		tokens := utils.ManagerUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

		dt := strfmt.DateTime(time.Now())
		firstOrderStatus := orders.NewAddNewOrderStatusParamsWithContext(ctx)
		firstOrderStatus.Data = &models.NewOrderStatus{
			OrderID:   firstOrderID,
			CreatedAt: &dt,
			Status:    &domain.OrderStatusApproved,
			Comment:   &domain.OrderStatusApproved,
		}
		os1, err := client.Orders.AddNewOrderStatus(firstOrderStatus, auth)
		require.NoError(t, err)
		require.NotNil(t, os1)

		secondOrderStatus := orders.NewAddNewOrderStatusParamsWithContext(ctx)
		secondOrderStatus.Data = &models.NewOrderStatus{
			OrderID:   secondOrderID,
			CreatedAt: &dt,
			Status:    &domain.OrderStatusRejected,
			Comment:   &domain.OrderStatusRejected,
		}
		os2, err := client.Orders.AddNewOrderStatus(secondOrderStatus, auth)
		require.NoError(t, err)
		require.NotNil(t, os2)

		params := equipment.NewBlockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		params.Data = &models.ChangeEquipmentStatusToBlockedRequest{
			StartDate: strfmt.DateTime(startDate),
			EndDate:   strfmt.DateTime(endDate),
		}

		res, err := client.Equipment.BlockEquipment(params, auth)
		require.NoError(t, err)
		require.True(t, res.IsCode(http.StatusNoContent))

		// The first order switches status from Approved to Blocked
		ok, err := checkOrderStatus(ctx, client, auth, firstOrderID, domain.OrderStatusBlocked)
		require.NoError(t, err)
		require.True(t, ok)

		// The second order keeps the same status
		ok, err = checkOrderStatus(ctx, client, auth, secondOrderID, domain.OrderStatusRejected)
		require.NoError(t, err)
		require.True(t, ok)

		unavailParams := equipment.NewGetUnavailabilityPeriodsByEquipmentIDParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)

		resp, err := client.Equipment.GetUnavailabilityPeriodsByEquipmentID(unavailParams, auth)
		require.NoError(t, err)
		require.True(t, resp.IsCode(http.StatusOK))

		dsG, err := time.Parse(time.RFC3339Nano, startDate.String())
		require.NoError(t, err)
		deG, err := time.Parse(time.RFC3339Nano, endDate.String())
		require.NoError(t, err)

		valStartDate, err := time.Parse(time.RFC3339Nano, resp.Payload.Items[1].StartDate.String())
		require.NoError(t, err)
		valEndDate, err := time.Parse(time.RFC3339Nano, resp.Payload.Items[1].EndDate.String())
		require.NoError(t, err)

		require.Equal(t, true,
			(valStartDate.Year() == dsG.Year()) &&
				(valStartDate.Month() == dsG.Month()) &&
				(valStartDate.Day() == dsG.Day()),
		)
		require.Equal(t, true,
			(valEndDate.Year() == deG.Year()) &&
				(valEndDate.Month() == deG.Month()) &&
				(valEndDate.Day() == deG.Day()),
		)
	})

	t.Run("Block Equipment is failed, equipment not found", func(t *testing.T) {
		tokens := utils.ManagerUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
		var fakeID int64 = -1

		params := equipment.NewBlockEquipmentParamsWithContext(ctx).WithEquipmentID(fakeID)
		params.Data = &models.ChangeEquipmentStatusToBlockedRequest{
			StartDate: strfmt.DateTime(startDate),
			EndDate:   strfmt.DateTime(endDate),
		}

		_, err := client.Equipment.BlockEquipment(params, auth)
		require.Error(t, err)

		wantErr := equipment.NewBlockEquipmentNotFound()
		codeExp := int32(http.StatusNotFound)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrEquipmentNotFound,
		}
		assert.Equal(t, wantErr, err)
	})

	t.Run("Block Equipment is prohibited for operators", func(t *testing.T) {
		tokens := utils.OperatorUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
		params := equipment.NewBlockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		params.Data = &models.ChangeEquipmentStatusToBlockedRequest{
			StartDate: strfmt.DateTime(startDate),
			EndDate:   strfmt.DateTime(endDate),
		}

		_, err = client.Equipment.BlockEquipment(params, auth)
		require.Error(t, err)

		wantErr := equipment.NewBlockEquipmentDefault(http.StatusForbidden)
		codeExp := int32(http.StatusForbidden)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrEquipmentBlockForbidden,
		}
		assert.Equal(t, wantErr, err)
	})

	t.Run("Block Equipment is prohibited for admins", func(t *testing.T) {
		tokens := utils.AdminUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
		params := equipment.NewBlockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		params.Data = &models.ChangeEquipmentStatusToBlockedRequest{
			StartDate: strfmt.DateTime(startDate),
			EndDate:   strfmt.DateTime(endDate),
		}

		_, err = client.Equipment.BlockEquipment(params, auth)
		require.Error(t, err)

		wantErr := equipment.NewBlockEquipmentDefault(http.StatusForbidden)
		codeExp := int32(http.StatusForbidden)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrEquipmentBlockForbidden,
		}
		assert.Equal(t, wantErr, err)
	})

	t.Run("Block Equipment is permitted for managers", func(t *testing.T) {
		tokens := utils.ManagerUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
		params := equipment.NewBlockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		params.Data = &models.ChangeEquipmentStatusToBlockedRequest{
			StartDate: strfmt.DateTime(startDate),
			EndDate:   strfmt.DateTime(endDate),
		}

		res, err := client.Equipment.BlockEquipment(params, auth)
		require.NoError(t, err)
		require.True(t, res.IsCode(http.StatusNoContent))
	})

	t.Run("Block Equipment is failed, endDate before startDate", func(t *testing.T) {
		tokens := utils.ManagerUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
		params := equipment.NewBlockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		params.Data = &models.ChangeEquipmentStatusToBlockedRequest{
			StartDate: strfmt.DateTime(endDate),
			EndDate:   strfmt.DateTime(startDate),
		}

		_, err := client.Equipment.BlockEquipment(params, auth)
		require.Error(t, err)

		wantErr := equipment.NewBlockEquipmentDefault(http.StatusBadRequest)
		codeExp := int32(http.StatusBadRequest)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrStartDateAfterEnd,
		}
		assert.Equal(t, wantErr, err)
	})

	t.Run("Update Block Equipment period", func(t *testing.T) {
		tokens := utils.ManagerUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
		updateStartDate, updateEndDate := time.Now().AddDate(0, 0, 3), time.Now().AddDate(0, 0, 14)

		params := equipment.NewBlockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		params.Data = &models.ChangeEquipmentStatusToBlockedRequest{
			StartDate: strfmt.DateTime(updateStartDate),
			EndDate:   strfmt.DateTime(updateEndDate),
		}

		res, err := client.Equipment.BlockEquipment(params, auth)
		require.NoError(t, err)
		require.True(t, res.IsCode(http.StatusNoContent))

		unavailParams := equipment.NewGetUnavailabilityPeriodsByEquipmentIDParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)

		resp, err := client.Equipment.GetUnavailabilityPeriodsByEquipmentID(unavailParams, auth)
		require.NoError(t, err)
		require.True(t, resp.IsCode(http.StatusOK))

		dsG, err := time.Parse(time.RFC3339Nano, startDate.String())
		require.NoError(t, err)
		deG, err := time.Parse(time.RFC3339Nano, endDate.String())
		require.NoError(t, err)

		valStartDate, err := time.Parse(time.RFC3339Nano, resp.Payload.Items[1].StartDate.String())
		require.NoError(t, err)
		valEndDate, err := time.Parse(time.RFC3339Nano, resp.Payload.Items[1].EndDate.String())
		require.NoError(t, err)

		require.Equal(t, true,
			(valStartDate.Year() == dsG.Year()) &&
				(valStartDate.Month() == dsG.Month()) &&
				(valStartDate.Day() == dsG.Day()),
		)
		require.Equal(t, true,
			(valEndDate.Year() == deG.Year()) &&
				(valEndDate.Month() == deG.Month()) &&
				(valEndDate.Day() == deG.Day()),
		)
	})
}

func TestIntegration_GetEquipments_WithBlockingPeriods(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	tokens := utils.ManagerUserLogin(t)

	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)
	equip, err := createEquipment(ctx, client, auth, model)
	require.NotNil(t, equip)
	require.NoError(t, err)

	blockParams := equipment.NewBlockEquipmentParamsWithContext(ctx).WithEquipmentID(*equip.Payload.ID)
	startDate1 := strfmt.DateTime(strfmt.DateTime(time.Now().AddDate(0, 0, 1)))
	endDate1 := strfmt.DateTime(time.Now().AddDate(0, 0, 11))
	blockParams.Data = &models.ChangeEquipmentStatusToBlockedRequest{
		StartDate: startDate1,
		EndDate:   endDate1,
	}
	_, err = client.Equipment.BlockEquipment(blockParams, auth)
	require.NoError(t, err)
	startDate2 := strfmt.DateTime(strfmt.DateTime(time.Now().AddDate(0, 1, 0)))
	endDate2 := strfmt.DateTime(time.Now().AddDate(0, 1, 10))
	blockParams.Data = &models.ChangeEquipmentStatusToBlockedRequest{
		StartDate: startDate2,
		EndDate:   endDate2,
	}
	_, err = client.Equipment.BlockEquipment(blockParams, auth)
	require.NoError(t, err)

	t.Run("Get All Equipment", func(t *testing.T) {
		params := equipment.NewGetAllEquipmentParamsWithContext(ctx)
		res, err := client.Equipment.GetAllEquipment(params, auth)
		require.NoError(t, err)
		for _, eq := range res.Payload.Items {
			if *eq.ID == *equip.Payload.ID {
				require.Equal(t, 1, len(eq.BlockingPeriods))
			}
		}
	})

	t.Run("Get Equipment by ID with block periods", func(t *testing.T) {
		params := equipment.NewGetEquipmentParamsWithContext(ctx)
		params.EquipmentID = *equip.Payload.ID
		res, err := client.Equipment.GetEquipment(params, auth)
		require.NoError(t, err)
		require.Equal(t, 1, len(res.Payload.BlockingPeriods))
	})
}

func TestIntegration_UnblockEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	tokens := utils.AdminUserLogin(t)
	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)
	eq, err := createEquipment(ctx, client, auth, model)
	require.NotNil(t, eq)
	require.NoError(t, err)

	t.Run("Unblock Equipment is prohibited for operators", func(t *testing.T) {
		tokens := utils.OperatorUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

		params := equipment.NewUnblockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		_, err = client.Equipment.UnblockEquipment(params, auth)
		require.Error(t, err)

		wantErr := equipment.NewUnblockEquipmentDefault(http.StatusForbidden)
		codeExp := int32(http.StatusForbidden)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrEquipmentUnblockForbidden,
		}
		assert.Equal(t, wantErr, err)
	})

	t.Run("Unblock Equipment is prohibited for admins", func(t *testing.T) {
		tokens := utils.AdminUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

		params := equipment.NewUnblockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		_, err = client.Equipment.UnblockEquipment(params, auth)
		require.Error(t, err)

		wantErr := equipment.NewUnblockEquipmentDefault(http.StatusForbidden)
		codeExp := int32(http.StatusForbidden)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrEquipmentUnblockForbidden,
		}
		assert.Equal(t, wantErr, err)
	})

	t.Run("Unblock Equipment is permitted for managers", func(t *testing.T) {
		token := utils.ManagerUserLogin(t)
		auth := utils.AuthInfoFunc(token.GetPayload().AccessToken)

		startDate, endDate := strfmt.DateTime(time.Now().AddDate(0, 0, 1)), strfmt.DateTime(time.Now().AddDate(0, 0, 11))
		blockParams := equipment.NewBlockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		blockParams.Data = &models.ChangeEquipmentStatusToBlockedRequest{
			StartDate: strfmt.DateTime(startDate),
			EndDate:   strfmt.DateTime(endDate),
		}

		blockRes, err := client.Equipment.BlockEquipment(blockParams, auth)
		require.NoError(t, err)
		require.True(t, blockRes.IsCode(http.StatusNoContent))

		eqStatusIDBlocked, err := getEquipmentStatus(ctx, client, *eq.Payload.ID, auth)
		require.NoError(t, err)

		unblockParams := equipment.NewUnblockEquipmentParamsWithContext(ctx).WithEquipmentID(*eq.Payload.ID)
		unblockRes, err := client.Equipment.UnblockEquipment(unblockParams, auth)
		require.NoError(t, err)
		require.True(t, unblockRes.IsCode(http.StatusNoContent))

		eqStatusIDUnblocked, err := getEquipmentStatus(ctx, client, *eq.Payload.ID, auth)
		require.NoError(t, err)
		require.NotEqual(t, eqStatusIDUnblocked, eqStatusIDBlocked)

		orStartDate, orEndDate := time.Now(), time.Now().AddDate(0, 0, 2)
		firstOrderID, err := createOrder(ctx, client, auth, eq.Payload.ID, orStartDate, orEndDate)
		require.NoError(t, err)
		require.NotNil(t, firstOrderID)
	})

	t.Run("Unblock Equipment is failed, equipment not found", func(t *testing.T) {
		tokens := utils.ManagerUserLogin(t)
		auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)
		var fakeID int64 = -1

		params := equipment.NewUnblockEquipmentParamsWithContext(ctx).WithEquipmentID(fakeID)
		_, err := client.Equipment.UnblockEquipment(params, auth)
		require.Error(t, err)

		wantErr := equipment.NewUnblockEquipmentNotFound()
		codeExp := int32(http.StatusNotFound)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrEquipmentNotFound,
		}
		assert.Equal(t, wantErr, err)
	})
}

func TestIntegration_DeleteEquipment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	tokens := utils.AdminUserLogin(t)

	auth := utils.AuthInfoFunc(tokens.GetPayload().AccessToken)

	model, err := setParameters(ctx, client, auth)
	require.NoError(t, err)
	eq, err := createEquipment(ctx, client, auth, model)
	require.NotNil(t, eq)
	require.NoError(t, err)

	t.Run("Delete All Equipment", func(t *testing.T) {
		beforeEq, err := client.Equipment.GetAllEquipment(equipment.NewGetAllEquipmentParamsWithContext(ctx), auth)
		require.NoError(t, err)
		assert.NotZero(t, len(beforeEq.Payload.Items))

		params := equipment.NewDeleteEquipmentParamsWithContext(ctx)
		for _, item := range beforeEq.Payload.Items {
			params.WithEquipmentID(*item.ID)
			resp, err := client.Equipment.DeleteEquipment(params, auth)
			require.True(t, resp.IsCode(http.StatusOK))
			require.NoError(t, err)
		}
	})

	t.Run("Delete Equipment failed: zero equipments, delete failed", func(t *testing.T) {
		params := equipment.NewDeleteEquipmentParamsWithContext(ctx)
		params.WithEquipmentID(int64(1))
		_, gotErr := client.Equipment.DeleteEquipment(params, auth)
		require.Error(t, gotErr)

		wantErr := equipment.NewDeleteEquipmentDefault(500)
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrDeleteEquipment,
			Details: "ent: equipment not found",
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Equipment failed: auth failed", func(t *testing.T) {
		params := equipment.NewDeleteEquipmentParamsWithContext(ctx)
		params.WithEquipmentID(int64(1))
		token := utils.TokenNotExist

		_, gotErr := client.Equipment.DeleteEquipment(params, utils.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := equipment.NewDeleteEquipmentDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}

func createOrder(ctx context.Context, be *client.Be, auth runtime.ClientAuthInfoWriterFunc, id *int64, start time.Time, end time.Time) (*int64, error) {
	rentStart := strfmt.NewDateTime()
	dateTimeFmt := "2006-01-02 15:04:05"
	err := rentStart.UnmarshalText([]byte(start.Format(dateTimeFmt)))
	if err != nil {
		return nil, err
	}
	rentEnd := strfmt.NewDateTime()
	err = rentEnd.UnmarshalText([]byte(end.Format(dateTimeFmt)))
	if err != nil {
		return nil, err
	}

	orderCreated, err := be.Orders.CreateOrder(&orders.CreateOrderParams{
		Context: ctx,
		Data: &models.OrderCreateRequest{
			Description: "order",
			EquipmentID: id,
			RentEnd:     &rentEnd,
			RentStart:   &rentStart,
		},
	}, auth)

	if err != nil {
		return nil, err
	}
	return orderCreated.Payload.ID, nil
}

func checkOrderStatus(ctx context.Context, be *client.Be, auth runtime.ClientAuthInfoWriterFunc, orderId *int64,
	statusName string) (bool, error) {
	orders, err := be.Orders.GetOrdersByStatus(
		orders.NewGetOrdersByStatusParamsWithContext(ctx).WithStatus(statusName), auth)
	if err != nil {
		return false, err
	}
	for _, order := range orders.Payload.Items {
		if *order.ID == *orderId {
			return true, nil
		}
	}
	return false, nil
}

func setParameters(ctx context.Context, client *client.Be, auth runtime.ClientAuthInfoWriterFunc) (*models.Equipment, error) {
	termsOfUse := "https://..."
	cost := int64(3900)
	condition := "удовлетворительное, местами облупляется краска"
	description := "удобная, подойдет для котов любых размеров"
	inventoryNumber := int64(1)

	category, err := client.Categories.GetCategoryByID(categories.NewGetCategoryByIDParamsWithContext(ctx).WithCategoryID(1), auth)
	if err != nil {
		return nil, err
	}

	subCategory, err := client.Subcategories.GetSubcategoryByID(subcategories.NewGetSubcategoryByIDParamsWithContext(ctx).WithSubcategoryID(1), auth)
	if err != nil {
		return nil, err
	}

	location := int64(71)
	mdays := int64(10)
	catName := "Том"
	rDate := int64(1520294400)

	status, err := client.EquipmentStatusName.GetEquipmentStatusName(
		eqStatusName.NewGetEquipmentStatusNameParamsWithContext(ctx).WithStatusID(1), auth)
	if err != nil {
		return nil, err
	}

	f, err := os.Open("../common/cat.jpeg")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	petSize, err := client.PetSize.GetAllPetSize(pet_size.NewGetAllPetSizeParamsWithContext(ctx), auth)
	if err != nil {
		return nil, err
	}

	photo, err := client.Photos.CreateNewPhoto(photos.NewCreateNewPhotoParams().WithContext(ctx).WithFile(f), auth)
	if err != nil {
		return nil, err
	}

	cats, err := client.PetKind.GetPetKind(pet_kind.NewGetPetKindParamsWithContext(ctx).WithPetKindID(1), auth)
	if err != nil {
		return nil, err
	}

	supp := "ИП Григорьев Виталий Васильевич"
	techIss := false
	title := "клетка midwest icrate 1"

	var subCategoryInt64 int64
	if subCategory.Payload.Data.ID != nil {
		subCategoryInt64 = *subCategory.Payload.Data.ID
	}

	return &models.Equipment{
		TermsOfUse:       termsOfUse,
		CompensationCost: &cost,
		Condition:        condition,
		Description:      &description,
		InventoryNumber:  &inventoryNumber,
		Category:         category.Payload.Data.ID,
		Subcategory:      subCategoryInt64,
		Location:         &location,
		MaximumDays:      &mdays,
		Name:             &catName,
		NameSubstring:    "box",
		PetKinds:         []int64{*cats.Payload.ID},
		PetSize:          petSize.Payload[0].ID,
		PhotoID:          photo.Payload.Data.ID,
		ReceiptDate:      &rDate,
		Status:           &status.Payload.Data.ID,
		Supplier:         &supp,
		TechnicalIssues:  &techIss,
		Title:            &title,
	}, nil
}

func createEquipment(ctx context.Context, client *client.Be, auth runtime.ClientAuthInfoWriterFunc, model *models.Equipment) (*equipment.CreateNewEquipmentCreated, error) {
	paramsCreate := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
	paramsCreate.NewEquipment = model
	created, err := client.Equipment.CreateNewEquipment(paramsCreate, auth)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func getEquipmentStatus(ctx context.Context, client *client.Be, id int64, auth runtime.ClientAuthInfoWriterFunc) (int64, error) {
	params := equipment.NewGetEquipmentParamsWithContext(ctx)
	params.EquipmentID = id
	eq, err := client.Equipment.GetEquipment(params, auth)
	if err != nil {
		return 0, err
	}
	return *eq.Payload.Status, nil
}
