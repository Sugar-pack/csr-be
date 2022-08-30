package orders

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/equipment"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/kinds"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/pet_kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/pet_size"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/photos"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/status"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/integration-tests/common"
)

var (
	auth  runtime.ClientAuthInfoWriterFunc
	eq    *models.EquipmentResponse
	token *string
)

func TestIntegration_BeforeOrderSetup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	client := common.SetupClient()

	l, p, err := common.GenerateLoginAndPassword()
	require.NoError(t, err)

	_, err = common.CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	loginUser, err := common.LoginUser(ctx, client, l, p)
	require.NoError(t, err)

	token = loginUser.GetPayload().AccessToken
	auth = common.AuthInfoFunc(token)

	eq, err = createEquipment(ctx, client, auth)
	require.NoError(t, err)
}

func TestIntegration_CreateOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	client := common.SetupClient()

	t.Run("Create Order failed: quantity more than kind.MaxReservationUnits", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		quantity := int64(20)
		equipment := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		params.Data = &models.OrderCreateRequest{
			Equipment:   equipment,
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}

		_, gotErr := client.Orders.CreateOrder(params, auth)
		require.Error(t, gotErr)

		wantErr := orders.NewCreateOrderDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "at most 10 allowed"}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create Order failed: access", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		quantity := int64(1)
		equipment := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		params.Data = &models.OrderCreateRequest{
			Equipment:   equipment,
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}
		incorrectToken := common.TokenNotExist
		_, gotErr := client.Orders.CreateOrder(params, common.AuthInfoFunc(&incorrectToken))
		require.Error(t, gotErr)

		wantErr := orders.NewCreateOrderDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create Order failed: start date should be before end date", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		quantity := int64(1)
		equipment := eq.ID
		rentEnd := strfmt.DateTime(time.Now())
		rentStart := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		params.Data = &models.OrderCreateRequest{
			Equipment:   equipment,
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}
		_, gotErr := client.Orders.CreateOrder(params, auth)
		require.Error(t, gotErr)

		wantErr := orders.NewCreateOrderDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "start date should be before end date"}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create Order failed: small rent period", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		quantity := int64(1)
		equipment := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour))
		params.Data = &models.OrderCreateRequest{
			Equipment:   equipment,
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}
		_, gotErr := client.Orders.CreateOrder(params, auth)
		require.Error(t, gotErr)

		wantErr := orders.NewCreateOrderDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "small rent period"}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create Order failed: too big reservation period", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		quantity := int64(1)
		equipment := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 1000000))
		params.Data = &models.OrderCreateRequest{
			Equipment:   equipment,
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}
		_, gotErr := client.Orders.CreateOrder(params, auth)
		require.Error(t, gotErr)

		wantErr := orders.NewCreateOrderDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: &models.ErrorData{Message: "too big reservation period"}}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create Order failed: validation error, required field", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		quantity := int64(1)
		//equipment := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour))
		params.Data = &models.OrderCreateRequest{
			Equipment:   nil,
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}
		_, gotErr := client.Orders.CreateOrder(params, auth)
		require.Error(t, gotErr)

		wantErr := orders.NewCreateOrderDefault(http.StatusUnprocessableEntity)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create Order", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		quantity := int64(1)
		equipment := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		params.Data = &models.OrderCreateRequest{
			Equipment:   equipment,
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}
		res, err := client.Orders.CreateOrder(params, auth)
		require.NoError(t, err)

		assert.Equal(t, equipment, res.Payload.Equipment.ID)
		assert.Equal(t, desc, *res.Payload.Description)
		assert.Equal(t, quantity, *res.Payload.Quantity)
		rentEnd.Equal(*res.Payload.RentEnd)
		rentStart.Equal(*res.Payload.RentStart)
	})

	t.Run("Create Order failed: duplicate order", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		quantity := int64(1)
		equipment := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		params.Data = &models.OrderCreateRequest{
			Equipment:   equipment,
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}

		_, err := client.Orders.CreateOrder(params, auth)
		require.Error(t, err)
	})
}

func TestIntegration_GetAllOrders(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := common.SetupClient()

	t.Run("Get All Orders Ok", func(t *testing.T) {
		params := orders.NewGetAllOrdersParamsWithContext(ctx)
		_, err := client.Orders.GetAllOrders(params, auth)
		require.NoError(t, err)
	})

	t.Run("Get All Orders Ok limit", func(t *testing.T) {
		eq2, err := createEquipment(ctx, client, auth)
		require.NoError(t, err)

		eq3, err := createEquipment(ctx, client, auth)
		require.NoError(t, err)

		createParams := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		quantity := int64(1)
		equipment := eq2.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		createParams.Data = &models.OrderCreateRequest{
			Equipment:   equipment,
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}
		_, err = client.Orders.CreateOrder(createParams, auth)
		require.NoError(t, err)

		createParams.Data.Equipment = eq3.ID
		_, err = client.Orders.CreateOrder(createParams, auth)
		require.NoError(t, err)

		limit := int64(1)
		offset := int64(0)
		orderBy := utils.AscOrder
		orderColumn := order.FieldID

		params := orders.NewGetAllOrdersParamsWithContext(ctx)
		params.OrderBy = &orderBy
		params.Limit = &limit
		params.Offset = &offset
		params.OrderColumn = &orderColumn
		res, err := client.Orders.GetAllOrders(params, auth)
		require.NoError(t, err)

		assert.Equal(t, int(limit), len(res.Payload.Items))
	})

	t.Run("Get All Orders failed: access", func(t *testing.T) {
		params := orders.NewGetAllOrdersParamsWithContext(ctx)
		token := common.TokenNotExist
		_, gotErr := client.Orders.GetAllOrders(params, common.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := orders.NewGetAllOrdersDefault(http.StatusInternalServerError)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Get All Orders failed: wrong column to order by - validation error", func(t *testing.T) {
		params := orders.NewGetAllOrdersParamsWithContext(ctx)
		limit := int64(1)
		offset := int64(0)
		orderBy := utils.AscOrder
		orderColumn := "wrong"

		params.OrderBy = &orderBy
		params.Limit = &limit
		params.Offset = &offset
		params.OrderColumn = &orderColumn
		_, gotErr := client.Orders.GetAllOrders(params, auth)
		require.Error(t, gotErr)

		wantErr := orders.NewGetAllOrdersDefault(http.StatusUnprocessableEntity)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

}

func TestIntegration_UpdateOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := common.SetupClient()

	eq4, err := createEquipment(ctx, client, auth)
	require.NoError(t, err)

	createParams := orders.NewCreateOrderParamsWithContext(ctx)
	desc := "test description"
	quantity := int64(1)
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createParams.Data = &models.OrderCreateRequest{
		Equipment:   eq4.ID,
		Description: &desc,
		Quantity:    &quantity,
		RentStart:   &rentStart,
		RentEnd:     &rentEnd,
	}
	order, err := client.Orders.CreateOrder(createParams, auth)
	require.NoError(t, err)

	orderID := order.Payload.ID
	t.Run("Update Order", func(t *testing.T) {
		params := orders.NewUpdateOrderParamsWithContext(ctx)
		params.OrderID = *orderID
		desc = "new"
		params.Data = &models.OrderUpdateRequest{
			Description: &desc,
			Quantity:    &quantity,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}

		res, err := client.Orders.UpdateOrder(params, auth)
		require.NoError(t, err)

		assert.Equal(t, desc, *res.Payload.Description)
		assert.Equal(t, quantity, *res.Payload.Quantity)
		rentEnd.Equal(*res.Payload.RentEnd)
		rentStart.Equal(*res.Payload.RentStart)
	})
}

func createEquipment(ctx context.Context, client *client.Be, auth runtime.ClientAuthInfoWriterFunc) (*models.EquipmentResponse, error) {
	params := equipment.NewCreateNewEquipmentParamsWithContext(ctx)
	model, err := setParameters(ctx, client, auth)
	if err != nil {
		return nil, err
	}

	params.NewEquipment = model

	res, err := client.Equipment.CreateNewEquipment(params, auth)
	if err != nil {
		return nil, err
	}
	return res.Payload, nil
}

func setParameters(ctx context.Context, client *client.Be, auth runtime.ClientAuthInfoWriterFunc) (*models.Equipment, error) {
	category := "Клетки"
	cost := int64(3900)
	condition := "удовлетворительное, местами облупляется краска"
	description := "удобная, подойдет для котов любых размеров"
	inventoryNumber := int64(1)

	kind, err := client.Kinds.GetKindByID(kinds.NewGetKindByIDParamsWithContext(ctx).WithKindID(1), auth)
	if err != nil {
		return nil, err
	}

	location := int64(71)
	amount := int64(1)
	mdays := int64(10)
	catName := "Том"
	rDate := "2018"

	status, err := client.Status.GetStatus(status.NewGetStatusParamsWithContext(ctx).WithStatusID(1), auth)
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
	techIss := "нет"
	title := "клетка midwest icrate 1"

	return &models.Equipment{
		Category:         &category,
		CompensationСost: &cost,
		Condition:        condition,
		Description:      &description,
		InventoryNumber:  &inventoryNumber,
		Kind:             &kind.Payload.Data.ID,
		Location:         &location,
		MaximumAmount:    &amount,
		MaximumDays:      &mdays,
		Name:             &catName,
		NameSubstring:    "box",
		PetKinds:         []int64{*cats.Payload.ID},
		PetSize:          &petSize.Payload[0].ID,
		PhotoID:          &photo.Payload.Data.ID,
		ReceiptDate:      &rDate,
		Status:           status.Payload.Data.ID,
		Supplier:         &supp,
		TechnicalIssues:  &techIss,
		Title:            &title,
	}, nil
}
