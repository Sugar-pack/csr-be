package orders

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

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/subcategories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/order"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/categories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/equipment"
	eqStatusName "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/equipment_status_name"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/orders"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/pet_kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/pet_size"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/photos"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
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

	login := common.AdminUserLogin(t)
	token = login.GetPayload().AccessToken
	auth = common.AuthInfoFunc(token)

	var err error
	eq, err = createEquipment(ctx, client, auth)
	require.NoError(t, err)
}

func TestIntegration_CreateOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	ctx := context.Background()
	client := common.SetupClient()

	t.Run("Create Order failed: access", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		eqID := int64(1)
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		params.Data = &models.OrderCreateRequest{
			Description: desc,
			EquipmentID: &eqID,
			RentEnd:     &rentEnd,
			RentStart:   &rentStart,
		}
		incorrectToken := common.TokenNotExist
		_, gotErr := client.Orders.CreateOrder(params, common.AuthInfoFunc(&incorrectToken))
		require.Error(t, gotErr)

		wantErr := orders.NewCreateOrderDefault(http.StatusUnauthorized)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Create Order failed: start date should be before end date", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		eqID := eq.ID
		rentEnd := strfmt.DateTime(time.Now())
		rentStart := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		params.Data = &models.OrderCreateRequest{
			Description: desc,
			EquipmentID: eqID,
			RentEnd:     &rentEnd,
			RentStart:   &rentStart,
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
		eqID := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour))
		params.Data = &models.OrderCreateRequest{
			Description: desc,
			EquipmentID: eqID,
			RentEnd:     &rentEnd,
			RentStart:   &rentStart,
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
		eqID := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 1000000))
		params.Data = &models.OrderCreateRequest{
			Description: desc,
			EquipmentID: eqID,
			RentEnd:     &rentEnd,
			RentStart:   &rentStart,
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
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour))
		params.Data = &models.OrderCreateRequest{
			Description: desc,
			RentEnd:     &rentEnd,
			RentStart:   &rentStart,
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
		eqID := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		params.Data = &models.OrderCreateRequest{
			Description: desc,
			EquipmentID: eqID,
			RentEnd:     &rentEnd,
			RentStart:   &rentStart,
		}
		res, err := client.Orders.CreateOrder(params, auth)
		require.NoError(t, err)

		//assert.Equal(t, equipment, res.Payload.Equipments[0].ID)
		assert.Equal(t, desc, *res.Payload.Description)
		rentEnd.Equal(*res.Payload.RentEnd)
		rentStart.Equal(*res.Payload.RentStart)
	})

	t.Run("Create Order failed: duplicate order", func(t *testing.T) {
		params := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		eqID := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		params.Data = &models.OrderCreateRequest{
			Description: desc,
			EquipmentID: eqID,
			RentEnd:     &rentEnd,
			RentStart:   &rentStart,
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
	equip, err := createEquipment(ctx, client, auth)
	assert.NoError(t, err)

	t.Run("Get All Orders Ok", func(t *testing.T) {
		wantOrders := 1
		params := orders.NewGetAllOrdersParamsWithContext(ctx)
		res, err := client.Orders.GetAllOrders(params, auth)
		require.NoError(t, err)

		// check that it has one created order
		assert.Equal(t, wantOrders, len(res.GetPayload().Items))

		// create another order and check that get returns +1 order
		//eq2, err := createEquipment(ctx, client, auth)
		//require.NoError(t, err)

		createParams := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		eqID := equip.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		createParams.Data = &models.OrderCreateRequest{
			Description: desc,
			EquipmentID: eqID,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}

		_, err = client.Orders.CreateOrder(createParams, auth)
		require.NoError(t, err)

		// orders number changed
		wantOrders = 2
		res, err = client.Orders.GetAllOrders(params, auth)
		require.NoError(t, err)

		assert.Equal(t, wantOrders, len(res.GetPayload().Items))
	})

	t.Run("Get All Orders Ok limit", func(t *testing.T) {
		//eq2, err := createEquipment(ctx, client, auth)
		//require.NoError(t, err)
		//
		//eq3, err := createEquipment(ctx, client, auth)
		//require.NoError(t, err)

		createParams := orders.NewCreateOrderParamsWithContext(ctx)
		desc := "test description"
		eqID := eq.ID
		rentStart := strfmt.DateTime(time.Now())
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
		createParams.Data = &models.OrderCreateRequest{
			Description: desc,
			EquipmentID: eqID,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}
		//_, err = client.Orders.CreateOrder(createParams, auth)
		//require.NoError(t, err)
		//
		//createParams.Data.Equipment = eq3.ID
		//_, err = client.Orders.CreateOrder(createParams, auth)
		//require.NoError(t, err)

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

		wantErr := orders.NewGetAllOrdersDefault(http.StatusUnauthorized)
		wantErr.Payload = &models.Error{Data: nil}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Get All Orders failed: validation error", func(t *testing.T) {
		params := orders.NewGetAllOrdersParamsWithContext(ctx)
		limit := int64(1)
		offset := int64(0)
		orderBy := utils.AscOrder
		// only id and rent_start can be used
		orderColumn := order.FieldRentEnd

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

	t.Run("Get All Orders OK: rent_start column to order by", func(t *testing.T) {
		params := orders.NewGetAllOrdersParamsWithContext(ctx)
		limit := int64(1)
		offset := int64(0)
		orderBy := utils.AscOrder
		// rent_start and id can be used for orderColumn only
		orderColumn := "rent_start"

		params.OrderBy = &orderBy
		params.Limit = &limit
		params.Offset = &offset
		params.OrderColumn = &orderColumn
		_, err := client.Orders.GetAllOrders(params, auth)
		require.NoError(t, err)
	})
}

func TestIntegration_List_Filtered(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := common.SetupClient()
	equip, err := createEquipment(ctx, client, auth)
	assert.NoError(t, err)

	ordersToCreate := 6 // create 6 orders to cover all statuses and have 1 order for each status
	existingOrders := 2

	for i := 1; i <= ordersToCreate; i++ {
		createParams := orders.NewCreateOrderParamsWithContext(ctx)
		desc := fmt.Sprintf("order %v", i)
		eqID := equip.ID
		rentStart := strfmt.DateTime(time.Now().Add(time.Hour * time.Duration(2 * i) * 24))
		rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * time.Duration(2 * i + 1) * 24))
		createParams.Data = &models.OrderCreateRequest{
			Description: desc,
			EquipmentID: eqID,
			RentStart:   &rentStart,
			RentEnd:     &rentEnd,
		}
		_, err := client.Orders.CreateOrder(createParams, auth)
		require.NoError(t, err)
	}

	t.Run("Get Orders All Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusAll
		// filter 'all', get all 7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, ordersToCreate+existingOrders, len(res.GetPayload().Items))
		for _, o := range res.Payload.Items {
			assert.Equal(t, domain.OrderStatusInReview, *o.LastStatus.Status)
		}
	})

	t.Run("Get Orders Active Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusActive
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, ordersToCreate+existingOrders, len(res.GetPayload().Items))
		for _, o := range res.Payload.Items {
			assert.Equal(t, domain.OrderStatusInReview, *o.LastStatus.Status)
		}
	})

	t.Run("Get Orders Finished zero", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusFinished
		// filter 'finished', 0 orders (all of them are active)
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 0, len(res.GetPayload().Items))
	})

	listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
	res, err := client.Orders.GetAllOrders(listParams, auth)
	require.NoError(t, err)

	managerLogin := common.ManagerUserLogin(t)
	managerToken := managerLogin.GetPayload().AccessToken
	managerAuth := common.AuthInfoFunc(managerToken)
	// approve all except the last one and leave the 1st in 'in review'
	for i, o := range res.Payload.Items {
		if i == 0 {
			continue
		}
		var st string
		if i != len(res.Payload.Items)-1 {
			st = domain.OrderStatusApproved
		} else {
			st = domain.OrderStatusRejected
		}
		dt := strfmt.DateTime(time.Now())
		osp := orders.NewAddNewOrderStatusParamsWithContext(ctx)
		osp.Data = &models.NewOrderStatus{
			OrderID: o.ID,
			CreatedAt: &dt,
			Status: &st,
			Comment: &st,
		}
		_, err = client.Orders.AddNewOrderStatus(osp, managerAuth)
		require.NoError(t, err)
	}

	t.Run("Get Orders 7 Active Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusActive
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 7, len(res.GetPayload().Items))
	})

	t.Run("Get Orders 6 Approved Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusApproved
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 6, len(res.GetPayload().Items))
	})

	t.Run("Get Orders 1 In_Review Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusInReview
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 1, len(res.GetPayload().Items))
	})

	t.Run("Get Orders 1 Finished Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusFinished
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 1, len(res.GetPayload().Items))
	})

	t.Run("Get Orders 1 Rejected Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusRejected
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 1, len(res.GetPayload().Items))
	})

	t.Run("Get Orders 0 Closed Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusClosed
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 0, len(res.GetPayload().Items))
	})

	// Close 1st Order
	dt := strfmt.DateTime(time.Now())
	osp := orders.NewAddNewOrderStatusParamsWithContext(ctx)
	osp.Data = &models.NewOrderStatus{
		OrderID: res.Payload.Items[0].ID,
		CreatedAt: &dt,
		Status: &domain.OrderStatusClosed,
		Comment: &domain.OrderStatusClosed,
	}
	_, err = client.Orders.AddNewOrderStatus(osp, auth)
	require.NoError(t, err)

	t.Run("Get Orders 6 Active Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusActive
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 6, len(res.GetPayload().Items))
	})	

	t.Run("Get Orders 2 Finished Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusFinished
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 2, len(res.GetPayload().Items))
	})	

	t.Run("Get Orders 1 Closed Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusClosed
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 1, len(res.GetPayload().Items))
	})

	t.Run("Get Orders 0 In_Review Ok", func(t *testing.T) {
		listParams := orders.NewGetAllOrdersParamsWithContext(ctx)
		listParams.Status = &domain.OrderStatusInReview
		// filter 'active', still  7 (5+2) orders
		res, err := client.Orders.GetAllOrders(listParams, auth)
		require.NoError(t, err)
		assert.Equal(t, 0, len(res.GetPayload().Items))
	})
}

func TestIntegration_UpdateOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := common.SetupClient()
	equip, err := createEquipment(ctx, client, auth)
	assert.NoError(t, err)

	createParams := orders.NewCreateOrderParamsWithContext(ctx)
	desc := "test description"
	eqID := equip.ID
	rentStart := strfmt.DateTime(time.Now())
	rentEnd := strfmt.DateTime(time.Now().Add(time.Hour * 24))
	createParams.Data = &models.OrderCreateRequest{
		Description: desc,
		EquipmentID: eqID,
		RentEnd:     &rentEnd,
		RentStart:   &rentStart,
	}
	order, err := client.Orders.CreateOrder(createParams, auth)
	require.NoError(t, err)

	quantity := int64(1)
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
	termsOfUse := "https://..."
	cost := int64(3900)
	condition := "удовлетворительное, местами облупляется краска"
	description := "удобная, подойдет для котов любых размеров"
	inventoryNumber := int64(1)

	category, err := client.Categories.GetCategoryByID(categories.NewGetCategoryByIDParamsWithContext(ctx).WithCategoryID(1), auth)
	if err != nil {
		return nil, err
	}

	subCat, err := client.Subcategories.GetSubcategoryByID(subcategories.NewGetSubcategoryByIDParamsWithContext(ctx).WithSubcategoryID(2), auth)
	if err != nil {
		return nil, err
	}

	location := int64(71)
	mdays := int64(10)
	catName := "Том"
	rDate := int64(1520345133)

	status, err := client.EquipmentStatusName.
		GetEquipmentStatusName(eqStatusName.NewGetEquipmentStatusNameParamsWithContext(ctx).WithStatusID(1), auth)
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

	var subCatInt64 int64
	if subCat.Payload.Data.ID != nil {
		subCatInt64 = *subCat.Payload.Data.ID
	}

	return &models.Equipment{
		TermsOfUse:       termsOfUse,
		CompensationCost: &cost,
		Condition:        condition,
		Description:      &description,
		InventoryNumber:  &inventoryNumber,
		Category:         category.Payload.Data.ID,
		Subcategory:      subCatInt64,
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
